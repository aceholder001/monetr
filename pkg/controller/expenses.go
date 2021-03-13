package controller

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http"
	"strings"
	"time"
)

// @tag.name Expenses
func (c *Controller) handleExpenses(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/expenses", c.getExpenses)
	p.Post("/{bankAccountId:uint64}/expenses", c.postExpenses)

	p.Put("/{bankAccountId:uint64}/expenses/{expenseId:uint64}", func(ctx *context.Context) {

	})

	p.Delete("/{bankAccountId:uint64}/expenses/{expenseId:uint64}", func(ctx *context.Context) {

	})
}

// List Expenses
// @id list-expenses
// @tags Expenses
// @description List all of the expenses for the specified bank account.
// @Security ApiKeyAuth
// @Param bankAccountId path int true "Bank Account ID"
// @Router /bank_accounts/{bankAccountId}/expenses [get]
// @Success 200 {array} models.Expense
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getExpenses(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	expenses, err := repo.GetExpenses(bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not retrieve expenses")
		return
	}

	ctx.JSON(expenses)
}

// Create Expense
// @id create-expense
// @tags Expense
// @summary Create an expense for the specified bank account.
// @security ApiKeyAuth
// @accept json
// @product json
// @Param bankAccountId path int true "Bank Account ID"
// @Param expense body models.Expense true "New Expense"
// @Router /bank_accounts/{bankAccountId}/expenses [post]
// @Success 200 {object} models.Expense
// @Failure 400 {object} InvalidBankAccountIdError "Invalid Bank Account ID."
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) postExpenses(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	var expense models.Expense
	if err := ctx.ReadJSON(&expense); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	expense.ExpenseId = 0 // Make sure we create a new expense.
	expense.BankAccountId = bankAccountId
	expense.Name = strings.TrimSpace(expense.Name)
	expense.Description = strings.TrimSpace(expense.Description)

	if expense.Name == "" {
		c.returnError(ctx, http.StatusBadRequest, "expense must have a name")
		return
	}

	expense.LastRecurrence = nil
	expense.NextRecurrence = expense.RecurrenceRule.After(time.Now(), false)

	if expense.FundingScheduleId != nil && *expense.FundingScheduleId > 0 {
		// Calculate the next contribution.
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.CreateExpense(&expense); err != nil {
		c.wrapPgError(ctx, err, "failed to create expense")
		return
	}

	ctx.JSON(expense)
}
