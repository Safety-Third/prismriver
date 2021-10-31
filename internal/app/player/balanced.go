package player

// InsertQueueItemBalanced inserts a given QueueItem into a given slice of QueueItems based on fairness, returning the
// result.
func InsertQueueItemBalanced(item *QueueItem, queue []*QueueItem) []*QueueItem {
	priority := make(map[uint32]uint64)
	for index, existing := range queue {
		if !existing.balanced {
			continue
		}
		if existingTotal, ok := priority[existing.owner]; ok {
			if itemTotal, ok := priority[item.owner]; ok {
				if existing.owner != item.owner && existingTotal > itemTotal  {
					queue = append(queue[:index + 1], queue[index:]...)
					queue[index] = item
					return queue
				} else {
					priority[existing.owner] += existing.Media.Length
				}
			} else {
				queue = append(queue[:index + 1], queue[index:]...)
				queue[index] = item
				return queue
			}
		} else {
			priority[existing.owner] = existing.Media.Length
		}
	}
	return append(queue, item)
}
