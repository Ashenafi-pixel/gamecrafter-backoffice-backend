package wait

import "sync"

// RoutineWaitGroup is set to handle graceful server shutdown.
//
// DO NOT reassign it!
var RoutineWaitGroup = &sync.WaitGroup{}
