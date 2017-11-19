package amper

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
)

type Kernel struct {
	sync.RWMutex

	containers      map[string]*Container
	policyTemplates map[string]*PolicyTemplate
	accounts        map[string]*Account

	StateBucket string
	S3          *s3.S3

	KeyFormat string
}

type AccountLimits struct {
	ManagedPolicySize      int
	ManagedPoliciesPerRole int
}

type Account struct {
	ID        string
	Name      string
	ShortName string

	Limits AccountLimits
}

func NewKernel(s3 *s3.S3, stateBucket string, keyFormat string) *Kernel {
	k := &Kernel{
		containers:      make(map[string]*Container),
		policyTemplates: make(map[string]*PolicyTemplate),
		accounts:        make(map[string]*Account),

		S3:          s3,
		StateBucket: stateBucket,
		KeyFormat:   keyFormat,
	}

	k.NewContainer("") // null container

	return k
}

func (a *Kernel) NewContainer(id string) (*Container, error) {
	a.Lock()
	defer a.Unlock()

	if _, ok := a.containers[id]; ok {
		return nil, fmt.Errorf("container '%s' already exists", id)
	}

	c := &Container{
		amper: a,
		ID:    id,
	}

	a.containers[id] = c

	return c, nil
}

func (a *Kernel) AddAccount(account *Account) error {
	a.Lock()
	defer a.Unlock()

	if _, ok := a.accounts[account.Name]; ok {
		return fmt.Errorf("account '%s' already exists", account.Name)
	}

	a.accounts[account.Name] = account

	if account.Limits.ManagedPolicySize == 0 {
		account.Limits.ManagedPolicySize = DefaultManagedPolicySize
	}

	if account.Limits.ManagedPoliciesPerRole == 0 {
		account.Limits.ManagedPoliciesPerRole = DefaultManagedPoliciesPerRole
	}

	return nil
}

func (a *Kernel) AddPolicyTemplate(containerID string, pt *PolicyTemplate) error {
	a.Lock()

	c, ok := a.containers[containerID]

	a.Unlock()

	if !ok {
		return fmt.Errorf("container '%s' not found", containerID)
	}

	if containerID == "" && pt.Template == nil {
		return fmt.Errorf("cannot add policy tmeplate, both contianerId and Template are not set")
	}

	return c.AddPolicyTemplate(pt)
}
