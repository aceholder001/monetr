package swag

import (
	"github.com/monetrapp/rest-api/pkg/models"
	"time"
)

type AlwaysSpending struct {
	// The desired funding schedule of the spending. Changing this will trigger a recalculation of the spending object.
	FundingScheduleId uint64 `json:"fundingScheduleId" example:"8539"`
	// Human friendly name of the spending object. Something like "Amazon Prime". But can be anything.
	Name string `json:"name" example:"Amazon Prime"`
	// Currently used as a description of the recurrence rule so that it does not need to be "generated" with each
	// pattern. This is not intended to be used by the end user and is generated by the UI when the spending object is
	// created or updated. However it can be modified if you were to send this request manually. It has no side affects,
	// it is simply used to better display data to the end user at this time.
	Description string `json:"description" example:"1st of every month"`
	// How much the spending object should allocate by the next recurrence date. For goals this target is reached once
	// and is considered complete, even if part of the total amount has been spent. For expenses this amount is
	// attempted to be allocated before the recurrence date regardless of spending. This means that even if a
	// transaction is spent from this spending object the allocation system will still allocate more funds to this
	// expense if the transaction was spent before it is technically due AND the funding schedule occurs before the
	// specified next recurrence date. Changing this amount will recalculate contributions to this spending object.
	TargetAmount int64 `json:"targetAmount" example:"1395"`
	// Recurrence rule telling the budgeting system how often this expense should be used. This helps the budgeting
	// system recalculate the next recurrence date each time an expense's recurrence date is reached. More information
	// about the format of the rule can be found here: https://tools.ietf.org/html/rfc5545
	// Note: These rules should be provided with the `RRULE:` prefix omitted if the tool you are using to generate the
	// rule strings include it. These rules are parsed using this library: https://github.com/teambition/rrule-go
	// Changing this rule would recalculate contributions to this spending object.
	RecurrenceRule *models.Rule `json:"recurrenceRule" swaggertype:"string" example:"FREQ=MONTHLY;BYMONTHDAY=1"`
	// The next time this expense or goal is due. For expenses this date is recalculated each time this date passes.
	// For goals this date is somewhat static. It can be modified but is not automatically recalculated once it is
	// reached. Changing this date would recalculate contributions to this spending object. These dates should be
	// provided in RFC3339 format with the timezone of the client included. The timezone is important as its used to
	// calculate the next time this expense recurs.
	NextRecurrence time.Time `json:"nextRecurrence" example:"2021-05-01T00:00:00-05:00"`
	// Indicate whether or not this spending object should receive contributions on it's funding schedule occurrence. If
	// the spending object is paused, the next time its funding schedule occurs, no additional amount will be allocated
	// to this spending object.
	IsPaused bool `json:"isPaused"`
}

type UpdateSpendingRequest struct {
	AlwaysSpending
	// The spending Id of the goal or expense that you are updating.
	SpendingId uint64 `json:"spendingId" example:"4364"`
}

type NewSpendingRequest struct {
	AlwaysSpending
	// Indicates which bank account the spending object is associated with. All spending objects must be associated with
	// one bank account. This value cannot be changed. It can only be set when the spending object is created.
	BankAccountId uint64 `json:"bankAccountId" example:"8437"`
	// The type of spending object this is. This cannot be changed. It can only be set when the spending object is created.
	// * 0 - Expense, the object will occur on a regular basis based on its recurrence rule. Spending from an expense will always change its next allocation amount.
	// * 1 - Goal, the object will allocate until it reaches it's target value and then stop. It can be spent from while it is still incomplete without changing the allocation amount.
	SpendingType models.SpendingType `json:"spendingType" example:"0" enums:"0,1"`
}

type SpendingResponse struct {
	NewSpendingRequest
	// The amount that has been allocated to the spending object. This amount is deducted from the available balance of
	// the bank account the spending object is associated with. It can be modified by spending a transaction from a
	// spending object. Or by transferring/allocating funds to a spending object. It cannot be modified directly.
	CurrentAmount int64 `json:"currentAmount" example:"1395"`
	// Used amount is only valid for goals at this time. It indicates how much has been spent from the spending object,
	// and is used to keep track of the goal's progress to its target without affecting the accuracy of the
	// `currentAmount` field. A goal is complete when the `currentAmount` + `usedAmount` = `targetAmount`.
	UsedAmount int64 `json:"usedAmount" example:"1043"`
	// The last time this spending object reset. A spending object is reset each time its `nextRecurrence` date elapses,
	// the `nextRecurrence` date is then moved to this field. This field is null if a spending object has never elapsed
	// before. Or if the spending object is a goal. This field is maintained automatically and cannot be modified.
	LastRecurrence *time.Time `json:"lastRecurrence" example:"2021-04-15T00:00:00-05:00"`
	// If a spending object cannot reach its `targetAmount` by the date that it is due on its funding schedule alone,
	// then the spending object is marked as "behind". This means that without manually transferring funds to the
	// spending object it will not have enough funds to fulfill its target by the due date. This value is calculated
	// automatically and cannot be changed.
	IsBehind bool `json:"isBehind" example:"false"`
	// When the spending object was initially created. This value cannot be changed.
	DateCreated time.Time `json:"dateCreated" example:"2021-04-04T12:43:23-05:00"`
}

type TransferResponse struct {
	// The balance of the bank account after the transferred allocations have been recalculated.
	Balance BalanceResponse `json:"balance"`
	// An array of spending objects that were updated during the transfer. By persisting these to the client's memory
	// the state of the spending objects is properly maintained.
	Spending []SpendingResponse `json:"spending"`
}
