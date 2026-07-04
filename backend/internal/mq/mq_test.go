package mq

import "testing"

func TestTopologyContainsRequiredBindings(t *testing.T) {
	topology := Topology()

	for _, name := range []string{"forum.events", "ai.events", "notification.events", "dead.exchange"} {
		if _, ok := topology.Exchange(name); !ok {
			t.Fatalf("missing exchange %s", name)
		}
	}
	for _, binding := range []Binding{
		{Exchange: ExchangeForumEvents, Queue: QueuePostTagging, RoutingKey: "post.created"},
		{Exchange: ExchangeForumEvents, Queue: QueueSearchIndex, RoutingKey: "post.*"},
		{Exchange: ExchangeForumEvents, Queue: QueueSearchIndex, RoutingKey: "comment.*"},
		{Exchange: ExchangeForumEvents, Queue: QueueAuditLog, RoutingKey: "post.*"},
		{Exchange: ExchangeAIEvents, Queue: QueueSearchIndex, RoutingKey: "ai.reply.*"},
		{Exchange: ExchangeForumEvents, Queue: QueueNotification, RoutingKey: "comment.created"},
		{Exchange: ExchangeForumEvents, Queue: QueueNotification, RoutingKey: "user.mentioned"},
		{Exchange: ExchangeAIEvents, Queue: QueueNotification, RoutingKey: "ai.reply.*"},
	} {
		if !topology.HasBinding(binding) {
			t.Fatalf("missing binding %#v", binding)
		}
	}
}
