package player

// InsertQueueItemDefault inserts a given QueueItem into a given slice of QueueItems at the bottom, returning the result.
func InsertQueueItemDefault(item *QueueItem, queue []*QueueItem) []*QueueItem {
	return append(queue, item)
}
