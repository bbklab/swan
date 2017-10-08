package mesos

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
	"github.com/Dataman-Cloud/swan/mesosproto"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

// StopKvmTask send message to kvm executor to stop kvm task
func (s *Scheduler) StopKvmTask(taskId, agentId, executorId string) error {
	return s.controlKvmTask(taskId, agentId, executorId, "Stop")
}

// StartKvmTask send message to kvm executor to start kvm task
func (s *Scheduler) StartKvmTask(taskId, agentId, executorId string) error {
	return s.controlKvmTask(taskId, agentId, executorId, "Start")
}

// SuspendKvmTask send message to kvm executor to suspend kvm task
func (s *Scheduler) SuspendKvmTask(taskId, agentId, executorId string) error {
	return s.controlKvmTask(taskId, agentId, executorId, "Suspend")
}

// ResumeKvmTask send message to kvm executor to resume kvm task
func (s *Scheduler) ResumeKvmTask(taskId, agentId, executorId string) error {
	return s.controlKvmTask(taskId, agentId, executorId, "Resume")
}

func (s *Scheduler) controlKvmTask(taskId, agentId, executorId, ops string) error {
	log.Printf("%sing kvm task %s with agentId %s", ops, taskId, agentId)

	if agentId == "" || executorId == "" {
		log.Warnf("agentId or executorId of task %s is empty, ignore", taskId)
		return nil
	}

	// TODO
	//t := NewTask(nil, taskId, "")
	//s.addPendingTask(t)
	//defer s.removePendingTask(taskId) // prevent leak

	var opcmd string
	switch ops {
	case "Start":
		opcmd = "SWAN_KVM_TASK_STARTUP"
	case "Stop":
		opcmd = "SWAN_KVM_TASK_SHUTDOWN"
	case "Suspend":
		opcmd = "SWAN_KVM_TASK_SUSPEND"
	case "Resume":
		opcmd = "SWAN_KVM_TASK_RESUME"
	default:
		return errors.New("unsupported kvm task operation")
	}

	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_MESSAGE.Enum(),
		Message: &mesosproto.Call_Message{
			ExecutorId: &mesosproto.ExecutorID{
				Value: proto.String(executorId),
			},
			AgentId: &mesosproto.AgentID{
				Value: proto.String(agentId),
			},
			Data: []byte(opcmd),
		},
	}

	// send call
	if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
		log.Errorf("%s controlKvmTask().SendCall() error: %v", ops, err)
		return err
	}

	// TODO
	//log.Debugf("Waiting for kvm task %s to be stopped by mesos", taskId)
	//for status := range t.GetStatus() {
	//log.Debugf("Receiving status %s for task %s", status.GetState().String(), taskId)
	//if IsKvmTaskStopped(status) {
	//log.Printf("Task %s stopped", taskId)
	//break
	//}
	//}

	return nil
}

// launch grouped runtime tasks with specified mesos offers
func (s *Scheduler) launchGroupKvmTasksWithOffers(offers []*magent.Offer, tasks []*Task) error {
	var (
		appId = strings.SplitN(tasks[0].GetName(), ".", 2)[1]
	)

	// build each tasks: set mesos agent id & build
	for _, task := range tasks {
		task.AgentId = &mesosproto.AgentID{
			Value: proto.String(offers[0].GetAgentId()),
		}
		task.Build()
	}

	// memo update each db tasks' AgentID, IP ...
	for _, t := range tasks {
		dbtask, err := s.db.GetKvmTask(appId, t.GetTaskId().GetValue())
		if err != nil {
			continue
		}

		dbtask.AgentId = t.AgentId.GetValue()
		dbtask.IPAddr = offers[0].GetHostname()
		dbtask.VncAddr = "" // TODO with ipaddr + vncport

		if err := s.db.UpdateKvmTask(appId, dbtask); err != nil {
			log.Errorln("update db kvm task got error: %v", err)
			continue
		}
	}

	// Construct Mesos Launch Call
	var (
		offerIds  = []*mesosproto.OfferID{}
		taskInfos = []*mesosproto.TaskInfo{}
	)

	for _, offer := range offers {
		offerIds = append(offerIds, &mesosproto.OfferID{
			Value: proto.String(offer.GetId()),
		})
	}

	for _, task := range tasks {
		taskInfos = append(taskInfos, &task.TaskInfo)
	}

	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_ACCEPT.Enum(),
		Accept: &mesosproto.Call_Accept{
			OfferIds: offerIds,
			Operations: []*mesosproto.Offer_Operation{
				&mesosproto.Offer_Operation{
					Type: mesosproto.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesosproto.Offer_Operation_Launch{
						TaskInfos: taskInfos,
					},
				},
			},
			Filters: &mesosproto.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	log.Printf("Launching %d kvm task(s) on agent %s", len(tasks), offers[0].GetHostname())

	// send call
	if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
		log.Errorln("launch().SendCall() error:", err)
		return fmt.Errorf("send launch call got error: %v", err)
	}

	return nil
}