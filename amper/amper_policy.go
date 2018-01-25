package amper

import (
	"encoding/json"
	"fmt"
)

type ServiceRolePolicy struct {
	Policy           *IAMPolicyDoc
	AssumeRolePolicy *IAMPolicyDoc
}

type Policy struct {
	amper *Kernel

	AccountPolicies map[string][]*IAMPolicyDoc

	ServiceRolePolicies map[string]map[string]*ServiceRolePolicy
}

const DefaultManagedPoliciesPerRole = 10
const DefaultManagedPolicySize = 6144

// @TODO implement proper bin-packing
func (policy *Policy) compressOne(account *Account, policies []*IAMPolicyDoc) ([]*IAMPolicyDoc, error) {
	var statements []*IAMPolicyStatement

	for _, p := range policies {
		for _, s := range p.Statements {
			s.size = s.Size()
		}
		statements = append(statements, p.Statements...)
	}

	var res []*IAMPolicyDoc

	for len(statements) > 0 {
		s := statements[0]

		for _, r := range res {
			if r.Size()+s.size+1 < account.Limits.ManagedPolicySize {
				r.Statements = append(r.Statements, s)
				s = nil
				statements = statements[1:]
				break
			} else if len(r.Statements) == 0 {
				// Rise error, if policy is not fitting in
				// empty policy document.
				return res, fmt.Errorf("Single policy statement is too big, cannot add to empty policy document, limit: %d, size: %d", account.Limits.ManagedPolicySize, r.Size()+s.size+1)
			}
		}

		if s != nil {
			if len(res) >= account.Limits.ManagedPoliciesPerRole {
				return res, fmt.Errorf("No enough space for policies in account %s, limit: %d", account.Name, account.Limits.ManagedPoliciesPerRole)
			}

			res = append(res, &IAMPolicyDoc{
				//@TODO set Id
				Version: IAMPolicyVersion,
			})
		}
	}

	return res, nil
}

func (p *Policy) compress() error {
	for account, policies := range p.AccountPolicies {
		policies, err := p.compressOne(p.amper.accounts[account], policies)

		if err != nil {
			return err
		}

		if policies == nil {
			// Attach empty policy
			policies = []*IAMPolicyDoc{{}}
		}

		p.AccountPolicies[account] = policies
	}

	return nil
}

// dump prints policy to stdout.
// For debugging purposes only.
func (p *Policy) dump() {
	for account, policies := range p.AccountPolicies {
		for k, p := range policies {
			s, err := json.MarshalIndent(p, "", "  ")

			if err != nil {
				panic(err)
			}

			fmt.Printf("Account=%s, PolicyIdx=%d, PolicyJSON=%s\n", account, k, s)
		}
	}
}
