package runnable

import "sigs.k8s.io/controller-runtime/pkg/manager"

type leaderElection struct {
	manager.Runnable
	needsLeaderElection bool
}

// LeaderElection wraps the given runnable to implement manager.LeaderElectionRunnable.
func LeaderElection(runnable manager.Runnable, needsLeaderElection bool) manager.Runnable {
	return &leaderElection{
		Runnable:            runnable,
		needsLeaderElection: needsLeaderElection,
	}
}

// RequireLeaderElection wraps the given runnable, marking it as not requiring leader election.
func NoLeaderElection(runnable manager.Runnable) manager.Runnable {
	return LeaderElection(runnable, false)
}

// NeedLeaderElection implements manager.NeedLeaderElection interface.
func (r *leaderElection) NeedLeaderElection() bool {
	return r.needsLeaderElection
}
