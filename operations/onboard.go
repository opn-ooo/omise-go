package operations

import "github.com/omise/omise-go/internal"

type CreateOnboard struct {
	Name              string   `schema:"name"`
	AgreementAccepted bool     `schema:"agreement_accepted"`
	DocumentIDs       []string `schema:"document_ids[]"`

	AccountDetail          CreateAccountDetail
	BusinessDetail         CreateBusinessDetail
	StateMentDetail        CreateStateMentDetail
	TransferDetail         CreateTransferDetail
	PolicyAcceptanceDetail CreatePolicyAcceptanceDetail
}

func (req *CreateOnboard) Describe() *internal.Description {
	return &internal.Description{
		Endpoint:    internal.APIStaging,
		Method:      "POST",
		Path:        "/onboard",
		ContentType: "multipart/form-data",
	}
}

type CreateAccountDetail struct {
	// Entity Type
	EntityType string `schema:"account_details[entity_type]"`

	// Company info (Company) / Business info (Individual)
	LegalName  string `schema:"account_details[legal_name]"`
	TaxID      string `schema:"account_details[tax_id]"`
	Address    string `schema:"account_details[address]"`
	BranchName string `schema:"account_details[branch]"`
	PostalCode string `schema:"account_details[postal_code]"`

	// Website
	WebsiteUrl   string `schema:"account_details[website_url]"`
	WebsiteNotes string `schema:"account_details[website_notes]"`

	// Signing Authority (Companies) / Contact (Individuals)
	FullName  string `schema:"account_details[full_name]"`
	BirthDate string `schema:"account_details[birth_date]"`
	Phone     string `schema:"account_details[phone]"`
	Mobile    string `schema:"account_details[mobile]"`
}

type CreateBusinessDetail struct {
	MerchantCategoryId     string `schema:"business_details[merchant_category_id]"`
	Description            string `schema:"business_details[description]"`
	OtherPayment           string `schema:"business_details[other_payment]"`
	BusinessAge            string `schema:"business_details[business_age]"`
	ApproximateTransaction string `schema:"business_details[approximate_transaction]"`
	BasketSize             string `schema:"business_details[basket_size]"`
	DeliveryMethod         string `schema:"business_details[delivery_method]"`
	RefundPolicy           string `schema:"business_details[refund_policy]"`
}

type CreateStateMentDetail struct {
	StatementName string `schema:"statement_details[statement_name]"`
}

type CreateTransferDetail struct {
	BankAccountBrand  string `schema:"transfer_details[bank_account_brand]"`
	BankAccountNumber string `schema:"transfer_details[bank_account_number]"`
	BankAccountName   string `schema:"transfer_details[bank_account_name]"`
}

type CreatePolicyAcceptanceDetail struct {
	TermsAndConditionsAccepted   bool `schema:"policy_acceptance_details[terms_and_conditions_accepted]"`
	PrivacyPolicyAccepted        bool `schema:"policy_acceptance_details[privacy_policy_accepted]"`
	DataProtectionPolicyAccepted bool `schema:"policy_acceptance_details[data_protection_policy_accepted]"`
	RefundPolicyAccepted         bool `schema:"policy_acceptance_details[refund_policy_accepted]"`
}
