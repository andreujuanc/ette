package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/itzmeanjan/ette/app/data"
	_db "github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest/graph/generated"
	"github.com/itzmeanjan/ette/app/rest/graph/model"
)

func (r *queryResolver) BlockByHash(ctx context.Context, hash string) (*model.Block, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	var block data.Block

	if res := db.Model(&_db.Blocks{}).Where("hash = ?", hash).First(&block).Error; res != nil {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleBlock(&block), nil
}

func (r *queryResolver) BlockByNumber(ctx context.Context, number string) (*model.Block, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	var block data.Block

	if res := db.Model(&_db.Blocks{}).Where("number = ?", _number).First(&block).Error; res != nil {
		return nil, errors.New("Bad Block Number")
	}

	return getGraphQLCompatibleBlock(&block), nil
}

func (r *queryResolver) BlockByNumberRange(ctx context.Context, from string, to string) ([]*model.Block, error) {
	_from, err := strconv.ParseUint(from, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Starting Block Number")
	}

	_to, err := strconv.ParseUint(to, 10, 64)
	if err != nil {
		return nil, errors.New("Bad End Block Number")
	}

	if !(_to-_from < 10) {
		return nil, errors.New("Block Number Range Too Large")
	}

	var blocks []*data.Block

	if res := db.Model(&_db.Blocks{}).Where("number >= ? and number <= ?", _from, _to).Order("number asc").Find(&blocks); res.Error != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleBlocks(blocks), nil
}

func (r *queryResolver) TransactionsByBlockHash(ctx context.Context, hash string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	var tx []*data.Transaction

	if res := db.Model(&_db.Transactions{}).Where("blockhash = ?", hash).Find(&tx); res.Error != nil {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleTransactions(tx), nil
}

func (r *queryResolver) TransactionsByBlockNumber(ctx context.Context, number string) ([]*model.Transaction, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	var tx []*data.Transaction

	if res := db.Model(&_db.Transactions{}).Where("blockhash = (?)", db.Model(&_db.Blocks{}).Where("number = ?", _number).Select("hash")).Find(&tx); res.Error != nil {
		return nil, errors.New("Bad Block Number")
	}

	return getGraphQLCompatibleTransactions(tx), nil
}

func (r *queryResolver) Transaction(ctx context.Context, hash string) (*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Transaction Hash")
	}

	var tx data.Transaction

	if err := db.Model(&_db.Transactions{}).Where("hash = ?", hash).First(&tx).Error; err != nil {
		return nil, errors.New("Bad Transaction Hash")
	}

	return getGraphQLCompatibleTransaction(&tx), nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
