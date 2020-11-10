# ette

Ethereum Blockchain Analyser 😎

## Table of Contents

- [Why did you build `ette` ?](#inspiration-)
- [What do I need to have to use it ?](#prerequisite-)
- [How to install it ?](#installation-)
- [How to use it ?](#usage-)

## Inspiration 🤔

I was looking for one tool which will be able to keep itself in sync with latest happenings on EVM based blockchain, while exposing REST API for querying blockchain data with various filters. That tool will also expose real time notification functionalities over websocket, when subscribed to topics.

It's not that I was unable find any solution, but wasn't fully satisfied with those, so I decided to write `ette`, which will do following

- Sync upto latest state of blockchain
- Listen for all happenings on EVM based blockchain
- Persist all happenings in local database
- Expose REST API for querying 👇, while also setting block range/ time range for filtering results. Allow querying latest **X** entries for events emitted by contracts.
    - Block data
    - Transaction data
    - Event data

- Expose websocket based real time notification mechanism for 
    - Blocks being mined
    - Transactions being sent from address and/ or transactions being received at address
    - Events being emitted by contract, with indexed fields i.e. topics

And here's `ette`

## Prerequisite 👍

- Make sure you've Go _( >= 1.15 )_ installed
- You need to also install & set up PostgreSQL. I found [this](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-20-04) guide helpful.
- Redis needs to be installed too. Consider following [this](https://www.digitalocean.com/community/tutorials/how-to-install-and-secure-redis-on-ubuntu-20-04) guide.
- Blockchain Node's websocket connection URL, required because we'll be listening for events in real time.

## Installation 🛠

- First fork this repository & clone it, some where out side of **GOPATH**.

```bash
git clone git@github.com:username/ette.git
```

- Now get inside `ette`

```bash
cd ette
```

- Create a `.env` file in this directory.

```
RPC=wss://<websocket-endpoint>
PORT=7000
DB_USER=user
DB_PASSWORD=password
DB_HOST=x.x.x.x
DB_PORT=5432
DB_NAME=ette
RedisConnection=tcp
RedisAddress=x.x.x.x:6379
```

- Now build `ette`

```bash
go build
```

- If everything goes as expected, you'll find one binary named, **ette** in this directory. Run it. 

```bash
./ette
```

- Database migration to be taken care of during application start up.
- Syncing `ette` with latest state of blockchain takes time. Whether syncing is done or not, can be checked by querying

```bash
curl http://localhost:7000/v1/synced
```

> Note: For production, you'll most probably run it using `systemd`

## Usage 🦾

`ette` exposes REST API for querying historical block, transaction & event related data. It can also play role of real time notification engine, when subscribed to supported topics.

### Historical Block Data 🤩

You can query historical block data with various combination of query string params. 👇 is a comprehensive guide for consuming block data.

URL | Method | Description
--- | --- | ---
`/v1/block?hash=0x...&tx=yes` | GET | Fetch all transactions present in a block, when block hash is known
`/v1/block?number=1&tx=yes` | GET | Fetch all transactions present in a block, when block number is known
`/v1/block?hash=0x...` | GET | Fetch block by hash
`/v1/block?number=1` | GET | Fetch block by number
`/v1/block?fromBlock=1&toBlock=10` | GET | Fetch blocks by block number range _( max 10 at a time )_
`/v1/block?fromTime=1604975929&toTime=1604975988` | GET | Fetch blocks by unix timestamp range _( max 60 seconds timespan )_

### Historical Transaction Data 😎

It's possible to query historical transactions data with various combination of query string params.

URL | Method | Description
--- | --- | ---
`/v1/transaction?hash=0x...` | GET | Fetch transaction by txHash
`/v1/transaction?nonce=1&fromAccount=0x...` | GET | Fetch transaction, when tx sender's address & account nonce are known
`/v1/transaction?fromBlock=1&toBlock=10&deployer=0x...` | GET | Find out what contracts are created by certain account within given block number range _( max 100 blocks )_
`/v1/transaction?fromTime=1604975929&toTime=1604975988&deployer=0x...` | GET | Find out what contracts are created by certain account within given timestamp range _( max 600 seconds of timespan )_
`/v1/transaction?fromBlock=1&toBlock=100&fromAccount=0x...&toAccount=0x...` | GET | Given block number range _( max 100 at a time )_ & a pair of accounts, can find out all tx performed between that pair, where `from` & `to` fields are fixed
`/v1/transaction?fromTime=1604975929&toTime=1604975988&fromAccount=0x...&toAccount=0x...` | GET | Given time stamp range _( max 600 seconds of timespan )_ & a pair of accounts, can find out all tx performed between that pair, where `from` & `to` fields are fixed
`/v1/transaction?fromBlock=1&toBlock=100&fromAccount=0x...` | GET | Given block number range _( max 100 at a time )_ & an account, can find out all tx performed from that account
`/v1/transaction?fromTime=1604975929&toTime=1604975988&fromAccount=0x...` | GET | Given time stamp range _( max 600 seconds of span )_ & an account, can find out all tx performed from that account
`/v1/transaction?fromBlock=1&toBlock=100&toAccount=0x...` | GET | Given block number range _( max 100 at a time )_ & an account, can find out all tx where target was this address
`/v1/transaction?fromTime=1604975929&toTime=1604975988&toAccount=0x...` | GET | Given time stamp range _( max 600 seconds of span )_ & an account, can find out all tx where target was this address

### Historical Event Data 🧐

`ette` lets you query historical event data, emitted by smart contracts, by combination of query string params.

URL | Method | Description
--- | --- | ---
`/v1/event?blockHash=0x...` | GET | Given blockhash, retrieves all events emitted by tx(s) present in block
`/v1/event?txHash=0x...` | GET | Given txhash, retrieves all events emitted during execution of this transaction
`/v1/event?count=50&contract=0x...` | GET | Returns last **x** _( <=50 )_ events emitted by this contract
`/v1/event?fromBlock=1&toBlock=10&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...&topic3=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0, 1, 2, 3}_
`/v1/event?fromBlock=1&toBlock=10&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0, 1, 2}_
`/v1/event?fromBlock=1&toBlock=10&contract=0x...&topic0=0x...&topic1=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0, 1}_
`/v1/event?fromBlock=1&toBlock=10&contract=0x...&topic0=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0}_
`/v1/event?fromBlock=1&toBlock=10&contract=0x...` | GET | Finding event(s) emitted from contract within given block range
`/v1/event?fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...&topic3=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0, 1, 2, 3}_
`/v1/event?fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0, 1, 2}_
`/v1/event?fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...&topic1=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0, 1}_
`/v1/event?fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0}_
`/v1/event?fromTime=1604975929&toTime=1604975988&contract=0x...` | GET | Finding event(s) emitted from contract within given time stamp range

**More coming soon**
