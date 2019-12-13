package payment_test

import (
	"github.com/onskycloud/go-redis"
	"github.com/onskycloud/payment"
	"testing"
	"time"
)

func TestContainer(t *testing.T) {
	cn := "123456"
	sn := "test"
	now := time.Now()
	options := Redis{
		Addr: "192.168.0.52:30333",
	}
	// model := &payment.SubscriptionCache{
	// 	UUID:             "123456",
	// 	PackageID:        "123456",
	// 	StartDate:        now,
	// 	EndDate:          now,
	// 	Type:             2,
	// 	Meta:             nil,
	// 	Duration:         0,
	// 	IntervalTime:     0,
	// 	Quota:            60,
	// 	OldPrice:        53500000,
	// 	CustomerNumber:   "123456",
	// 	Service:          "test",
	// 	Status:           2,
	// 	HaveTrialPackage: true,
	// 	TrialDuration:    10,
	// }
	model := &payment.SubscriptionCache{
		UUID:             "123456",
		PackageID:        "123456",
		StartDate:        now,
		EndDate:          now.Add(time.Hour * 10*24*30),
		Type:             3,
		Meta:             nil,
		Duration:         12*30,
		IntervalTime:     2,
		Quota:            60,
		OldPrice:        53500000,
		CustomerNumber:   "123456",
		Service:          "test",
		Status:           2,
		HaveTrialPackage: false,
		TrialDuration:    0,
	}
	client := redis.NewRedis(&options)

	client.SetObject("onsky:payment:subscriptions", cn+sn, model)

	testValidatePayment(client, cn, sn, t)
}
func testValidatePayment(db *redis.Redis, cn string, sn string, t *testing.T) {
	duration,err := payment.Validate(db, cn, sn)
	if err != nil {
		t.Fatalf("payment reject: %+v\n", err)
	}
	t.Logf("duration:%+v\n",duration)
}

// Redis configuration
type Redis struct {
	Addr     string `yaml:"addr,omitempty"`
	Password string `yaml:"password,omitempty"`
}
