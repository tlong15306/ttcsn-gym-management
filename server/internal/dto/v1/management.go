package v1dto

import (
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/member"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
)

type MemberRequest = member.Input
type MembershipRequest = member.MembershipInput
type RoomRequest = facility.RoomInput
type SlotRequest = facility.SlotInput
type SlotStatusRequest = facility.SlotStatusInput
type SessionRequest = facility.SessionInput
type TrainerProfileRequest = trainer.ProfileInput
type TrainerDecisionRequest = trainer.DecisionInput
type AvailabilityRequest = trainer.AvailabilityInput
type UnavailabilityRequest = trainer.Unavailability
