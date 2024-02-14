# GCrate

GCrate is a practical Golang toolkit that includes implementations based on efficiency optimization strategies, relevant data structures, and adds support for generics.

## Implementations:

### TimingWheel

Efficient delayed task scheduler for reducing frequent addition/removal of timers.

- [tickwheel]
- [delayedwheel]

### ShardMap

Avoid panics caused by concurrent read/write operations on maps while minimizing locking granularity to mitigate efficiency losses.

- [shardmap]

### DataStructure

- [list]
  - GenericList
  - SyncList

- [pqueue]
  - ProiorityQueue
  - DelayQueue