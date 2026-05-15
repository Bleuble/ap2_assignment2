package infrastructure

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type SimulatedProvider struct{}

func NewSimulatedProvider() *SimulatedProvider {
	rand.Seed(time.Now().UnixNano())
	return &SimulatedProvider{}
}

func (p *SimulatedProvider) SendEmail(to string, subject string, body string) error {
	log.Printf("[SimulatedProvider] Attempting to send email to %s...", to)

	time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(400)))

	if rand.Float32() < 0.20 {
		return fmt.Errorf("simulated network error or API timeout")
	}

	log.Printf("[SimulatedProvider] Successfully sent email to %s", to)
	return nil
}

type RealProvider struct{}

func NewRealProvider() *RealProvider {
	return &RealProvider{}
}

func (p *RealProvider) SendEmail(to string, subject string, body string) error {
	log.Printf("[RealProvider] Sending REAL email to %s via SMTP/Mailjet... (Not fully implemented in code, acting as dummy successful provider)", to)
	return nil
}
