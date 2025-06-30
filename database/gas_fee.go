package database

import (
	"gorm.io/gorm"
	"math/big"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type GasFee struct {
	GUID         uuid.UUID      `json:"guid" gorm:"primaryKey;DEFAULT replace(uuid_generate_v4()::text,'-','');serializer:uuid"`
	ChainId      *big.Int       `json:"chain_id" gorm:"serializer:u256"`
	TokenAddress common.Address `json:"token_address" gorm:"serializer:bytes"`
	GasFee       *big.Int       `json:"gas_fee" gorm:"serializer:u256"`
	Timestamp    uint64         `json:"timestamp"`
}

func (GasFee) TableName() string {
	return "gas_fee"
}

type gasFeeDB struct {
	gorm *gorm.DB
}

type GasFeeDB interface {
	GasFeeView
	StoreOrUpdateGasFee(msgHash *GasFee) error
}

type GasFeeView interface {
	QueryGasFees(chainId string) (*GasFee, error)
}

func NewGasFeeDB(db *gorm.DB) GasFeeDB {
	return &gasFeeDB{gorm: db}
}

func (db *gasFeeDB) StoreOrUpdateGasFee(gasFee *GasFee) error {
	var gasFeeRecord GasFee
	err := db.gorm.Table("gas_fee").Where("chain_id = ?", gasFee.ChainId.String()).Take(&gasFeeRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			result := db.gorm.Table("gas_fee").Omit("guid").Create(gasFee)
			return result.Error
		}
	}
	gasFeeRecord.GasFee = gasFee.GasFee
	gasFeeRecord.ChainId = gasFee.ChainId
	gasFeeRecord.Timestamp = gasFee.Timestamp
	err = db.gorm.Table("gas_fee").Save(gasFeeRecord).Error
	if err != nil {
		log.Error("save gas fee record fail", "err", err)
		return err
	}
	return nil
}

func (db *gasFeeDB) QueryGasFees(chainId string) (*GasFee, error) {
	var gasFee GasFee
	err := db.gorm.Table("gas_fee").Where("chain_id = ?", chainId).Take(&gasFee).Error
	if err != nil {
		log.Error("get gas fee fail", "err", err)
		return nil, err
	}
	return &gasFee, nil
}
