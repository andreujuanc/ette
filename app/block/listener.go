package block

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gammazero/workerpool"
	"github.com/gookit/color"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	q "github.com/itzmeanjan/ette/app/queue"
	"gorm.io/gorm"
)

// SubscribeToNewBlocks - Listen for event when new block header is
// available, then fetch block content ( including all transactions )
// in different worker
func SubscribeToNewBlocks(connection *d.BlockChainNodeConnection, _db *gorm.DB, status *d.StatusHolder, redis *d.RedisInfo, queue *q.BlockProcessorQueue) {
	headerChan := make(chan *types.Header)

	subs, err := connection.Websocket.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		log.Fatal(color.Red.Sprintf("[!] Failed to subscribe to block headers : %s", err.Error()))
	}
	// Scheduling unsubscribe, to be executed when end of this execution scope is reached
	defer subs.Unsubscribe()

	// Flag to check for whether this is first time block header being received or not
	//
	// If yes, we'll start syncer to fetch all block in range (last block processed, latest block)
	first := true
	// Creating a job queue of size `#-of CPUs present in machine`
	// where block fetching requests to be submitted
	wp := workerpool.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))
	// Scheduling worker pool closing, to be called,
	// when returning from this execution scope i.e. function
	defer wp.Stop()

	for {
		select {
		case err := <-subs.Err():

			log.Fatal(color.Red.Sprintf("[!] Listener stopped : %s", err.Error()))

		case header := <-headerChan:

			// At very beginning iteration, newly mined block number
			// should be greater than max block number obtained from DB
			if first && !(header.Number.Uint64() > status.MaxBlockNumberAtStartUp()) {

				log.Fatal(color.Red.Sprintf("[!] Bad block received : expected > `%d`\n", status.MaxBlockNumberAtStartUp()))

			}

			// At any iteration other than first one, if received block number
			// is more than latest block number + 1, it's definite that we've some
			// block (  >=1 ) missed & the RPC node we're relying on might be feeding us with
			// wrong data
			//
			// It's better stop relying on it, we crash the program
			// @note This is not the state-of-the art solution, but this is it, as of now
			// It can be improved.
			if !first && header.Number.Uint64() > status.GetLatestBlockNumber()+1 {

				log.Fatal(color.Red.Sprintf("[!] Bad block received %d, expected %d", header.Number.Uint64(), status.GetLatestBlockNumber()))

			}

			// At any iteration other than first one, if received block number
			// not exactly current latest block number + 1, then it's probably one
			// reorganization, we'll attempt to process this new block
			if !first && !(header.Number.Uint64() == status.GetLatestBlockNumber()+1) {

				log.Print(color.Blue.Sprintf("[*] Received block %d again, expected %d, attempting to process", header.Number.Uint64(), status.GetLatestBlockNumber()+1))

			} else {

				log.Print(color.Blue.Sprintf("[*] Received block %d, attempting to process", header.Number.Uint64()))

			}

			status.SetLatestBlockNumber(header.Number.Uint64())

			if first {

				// Starting now, to be used for calculating system performance, uptime etc.
				status.SetStartedAt()

				// Starting go routine for fetching blocks `ette` failed to process in previous attempt
				//
				// Uses Redis backed queue for fetching pending block hash & retries
				go RetryQueueManager(connection.RPC, _db, redis, queue, status)

				// If historical data query features are enabled
				// only then we need to sync to latest state of block chain
				if cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3" {

					// Starting syncer in another thread, where it'll keep fetching
					// blocks from highest block number it fetched last time to current network block number
					// i.e. trying to fill up gap, which was caused when `ette` was offline

					// Upper limit of syncing, in terms of block number
					from := header.Number.Uint64() - 1
					// Lower limit of syncing, in terms of block number
					//
					// Subtracting confirmation required block number count, due to
					// the fact it might be case those block contents might have changed due to
					// some reorg, in the time duration, when `ette` was offline
					//
					// So we've to take a look at those
					to := status.MaxBlockNumberAtStartUp() - cfg.GetBlockConfirmations()

					go SyncBlocksByRange(connection.RPC, _db, redis, queue, from, to, status)

				}
				// Making sure that when next latest block header is received, it'll not
				// start another syncer
				first = false

			}

			// As soon as new block is mined, `ette` will try to fetch it
			// and that job will be submitted in job queue
			//
			// Putting it in a different function scope for safety purpose
			// so that job submitter gets its own copy of block number & block hash,
			// otherwise it might get wrong info, if new block gets mined very soon &
			// this job is not yet submitted
			//
			// Though it'll be picked up sometime in future ( by missing block finder ), but it can be safely handled now
			// so that it gets processed immediately
			func(blockHash common.Hash, blockNumber uint64, _queue *q.BlockProcessorQueue) {

				// When only processing blocks in real-time mode
				// no need to check what's present in unfinalized block number queue
				// because no finality feature is provided for blocks on websocket based
				// real-time subscription mechanism
				if cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3" {

					// Attempting to submit all blocks to job processor queue
					// if more blocks are present in non-final queue, than actually
					// should be
					for GetUnfinalizedQueueLength(redis) > int64(cfg.GetBlockConfirmations()) {

						// Before submitting new block processing job
						// checking whether there exists any block in unfinalized
						// block queue or not
						//
						// If yes, we're attempting to process it, because it has now
						// achieved enough confirmations
						if CheckIfOldestBlockIsConfirmed(redis, status) {

							oldest := PopOldestBlockFromUnfinalizedQueue(redis)

							log.Print(color.Yellow.Sprintf("[*] Attempting to process finalised block %d [ Latest Block : %d | In Queue : %d ]", oldest, status.GetLatestBlockNumber(), GetUnfinalizedQueueLength(redis)))

							// Taking `oldest` variable's copy in local scope of closure, so that during
							// iteration over queue elements, none of them get missed, becuase we're
							// dealing with concurrent system, where previous `oldest` can be overwritten
							// by new `oldest` & we end up missing a block
							func(_oldestBlock uint64, _queue *q.BlockProcessorQueue) {

								wp.Submit(func() {

									if !_queue.Put(_oldestBlock) {
										return
									}

									if !FetchBlockByNumber(connection.RPC, _oldestBlock, _db, redis, false, queue, status) {

										_queue.Failed(_oldestBlock)
										return

									}

									_queue.Done(_oldestBlock)

								})

							}(oldest, _queue)

						} else {

							// If left most block is not yet finalized, it'll attempt to
							// reorganize that queue so that other blocks waiting to be processed
							// can get that opportunity
							//
							// This situation generally occurs due to concurrent pattern implemented
							// in block processor
							MoveUnfinalizedOldestBlockToEnd(redis)

						}

					}

				}

				wp.Submit(func() {

					if !_queue.Put(blockNumber) {
						return
					}

					if !FetchBlockByHash(connection.RPC, blockHash, fmt.Sprintf("%d", blockNumber), _db, redis, queue, status) {

						_queue.Failed(blockNumber)
						return

					}

					_queue.Done(blockNumber)

				})

			}(header.Hash(), header.Number.Uint64(), queue)

		}
	}
}
