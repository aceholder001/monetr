package jobs

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"time"
)

const (
	PullInitialTransactions = "PullInitialTransactions"
)

func (j *jobManagerBase) pullInitialTransactions(job *work.Job) error {
	span := sentry.StartSpan(context.Background(), "Job", sentry.TransactionName("Pull Initial Transactions"))
	defer span.Finish()

	log := j.getLogForJob(job)

	accountId := uint64(job.ArgInt64("accountId"))

	linkId := uint64(job.ArgInt64("linkId"))
	if linkId == 0 {
		log.Error("cannot pull initial transactions without a link Id")
		return errors.Errorf("cannot pull initial transactions without a link Id")
	}

	log = log.WithField("linkId", linkId)

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("cannot pull initial transactions for link provided")
			return nil
		}

		if link.LinkStatus == models.LinkStatusSetup {
			log.Warn("link has already been setup, initial transactions will not be pulled")
			return nil
		}

		if link.PlaidLink == nil {
			log.Error("provided link does not have any plaid credentials")
			return nil
		}

		if len(link.BankAccounts) == 0 {
			log.Error("no bank accounts for plaid link")
			return nil
		}

		bankAccountIdsByPlaid := map[string]uint64{}
		bankAccountIds := make([]string, len(link.BankAccounts))
		for i, bankAccount := range link.BankAccounts {
			bankAccountIds[i] = bankAccount.PlaidAccountId
			bankAccountIdsByPlaid[bankAccount.PlaidAccountId] = bankAccount.BankAccountId
		}

		now := time.Now().UTC()
		plaidTransactions, err := j.plaidClient.GetAllTransactions(
			span.Context(),
			link.PlaidLink.AccessToken,
			now.Add(-30*24*time.Hour),
			now,
			bankAccountIds,
		)

		if len(plaidTransactions) == 0 {
			log.Warn("no transactions were retrieved from plaid")
			return nil
		}

		log.Debugf("retreived %d transaction(s) from plaid, processing now", len(plaidTransactions))

		transactions := make([]models.Transaction, len(plaidTransactions))
		for i, plaidTransaction := range plaidTransactions {
			date, _ := time.Parse("2006-01-02", plaidTransaction.Date)
			var authorizedDate *time.Time
			if plaidTransaction.AuthorizedDate != "" {
				authDate, _ := time.Parse("2006-01-02", plaidTransaction.AuthorizedDate)
				authorizedDate = &authDate
			}
			transactions[i] = models.Transaction{
				AccountId:            repo.AccountId(),
				BankAccountId:        bankAccountIdsByPlaid[plaidTransaction.AccountID],
				PlaidTransactionId:   plaidTransaction.ID,
				Amount:               int64(plaidTransaction.Amount * 100),
				SpendingId:           nil,
				Categories:           plaidTransaction.Category,
				OriginalCategories:   plaidTransaction.Category,
				Date:                 date,
				AuthorizedDate:       authorizedDate,
				Name:                 plaidTransaction.Name,
				OriginalName:         plaidTransaction.Name,
				MerchantName:         plaidTransaction.MerchantName,
				OriginalMerchantName: plaidTransaction.MerchantName,
				IsPending:            plaidTransaction.Pending,
				CreatedAt:            now,
			}
		}

		// Reverse the list so the oldest records are inserted first.
		for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
			transactions[i], transactions[j] = transactions[j], transactions[i]
		}

		if err := repo.InsertTransactions(transactions); err != nil {
			log.WithError(err).Error("failed to store initial transactions")
			return err
		}

		link.LinkStatus = models.LinkStatusSetup
		if err = repo.UpdateLink(link); err != nil {
			log.WithError(err).Error("failed to update link status")
			return err
		}

		if err = j.ps.Notify(
			span.Context(),
			fmt.Sprintf("initial_plaid_link_%d_%d", accountId, link.LinkId),
			"success",
		); err != nil {
			log.WithError(err).Error("failed to publish link status to pubsub")
			return nil // Not good enough of a reason to fail.
		}

		return nil
	})
}
