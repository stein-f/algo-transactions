package main

import (
	"context"
	"math"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
)

const (
	startDate      = "2024-03-12T00:00:00Z"
	assetID        = 1387238831
	wallet         = "G6MVS6RZKTUZCP46O2WUTRP5FUFSETT5RUD2PZQ2SQWVSUW56D2Z7K435I"
	filterOnAmount = 2466
	decimals       = 6
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// create the indexer client
	idxClient, err := indexer.MakeClient("https://mainnet-idx.algonode.cloud", "")
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create indexer client")
	}

	// fetch transactions since startDate
	transactions, err := fetchTransactionsAfterTime(idxClient, wallet, assetID, startDate)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to fetch transactions")
	}

	log.Info().Msgf("found %d transactions since %s", len(transactions), startDate)

	// filter transactions by filterOnAmount
	var filteredTransactions []models.Transaction
	for _, txn := range transactions {
		if txn.AssetTransferTransaction.Amount == convertToFixedDecimal(filterOnAmount, decimals) {
			filteredTransactions = append(filteredTransactions, txn)
		}
	}

	log.Info().Msgf("found %d transactions with amount %d", len(filteredTransactions), filterOnAmount)

	var totalAmount uint64
	for _, txn := range filteredTransactions {
		totalAmount += txn.AssetTransferTransaction.Amount
	}

	log.Info().Msgf("total amount: %g", convertDecimal(totalAmount, decimals))
}

func fetchTransactionsAfterTime(indexerClient *indexer.Client, accountID string, assetID uint64, afterTime string) ([]models.Transaction, error) {
	next := ""
	var txns []models.Transaction
	ctx := context.Background()
	for {
		tx, err := indexerClient.LookupAccountTransactions(accountID).
			AfterTimeString(afterTime).
			Limit(1000).
			AssetID(assetID).
			NextToken(next).
			Do(ctx)
		if err != nil {
			return nil, err
		}
		next = tx.NextToken
		txns = append(txns, tx.Transactions...)
		if tx.NextToken == "" {
			break
		}
	}
	return txns, nil
}

func convertDecimal(number uint64, decimals int) float64 {
	divisor := math.Pow(10, float64(decimals))
	return float64(number) / divisor
}

func convertToFixedDecimal(number float64, decimals int) uint64 {
	multiplier := math.Pow(10, float64(decimals))
	return uint64(number * multiplier)
}
