package api

import (
	"github.com/Dataman-Cloud/swan/scheduler"
)

type Router struct {
	routes []Route
	sched  *scheduler.Scheduler
}

func NewRouter(sched *scheduler.Scheduler) *Router {
	r := &Router{
		sched: sched,
	}
	r.initRoutes()
	return r
}

func (r *Router) initRoutes() {
	r.routes = []Route{
		// task
		NewRoute("POST", "/tasks", r.tasksAdd),

		// app
		NewRoute("POST", "/v1/apps", r.applicationCreate),
		NewRoute("GET", "/v1/apps", r.applicationList),
		NewRoute("GET", "/v1/apps/{id}", r.applicationFetch),
		NewRoute("DELETE", "/v1/apps/{id}", r.applicationDelete),
	}
}
