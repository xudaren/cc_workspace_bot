package claude

import (
	"strings"
	"testing"

	"github.com/kid0317/cc-workspace-bot/internal/config"
)

func TestChannelKeyToRoutingKey(t *testing.T) {
	tests := []struct {
		channelKey string
		want       string
	}{
		{"p2p:ou_abc:cli_app1", "p2p:ou_abc"},
		{"group:oc_xyz:cli_app1", "group:oc_xyz"},
		{"thread:oc_xyz:tid_123:cli_app1", "group:oc_xyz"},
		{"", ""},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		got := channelKeyToRoutingKey(tt.channelKey)
		if got != tt.want {
			t.Errorf("channelKeyToRoutingKey(%q) = %q, want %q", tt.channelKey, got, tt.want)
		}
	}
}

func TestInjectRoutingContext(t *testing.T) {
	appCfg := &config.AppConfig{ID: "app1"}

	t.Run("injects routing block when channel key set", func(t *testing.T) {
		req := &ExecuteRequest{
			Prompt:     "hello",
			ChannelKey: "p2p:ou_abc:cli_app1",
			SenderID:   "ou_abc",
			AppConfig:  appCfg,
		}
		got := injectRoutingContext(req)
		if !strings.HasPrefix(got, "<system_routing>") {
			t.Errorf("expected <system_routing> prefix, got: %q", got)
		}
		if !strings.Contains(got, "routing_key: p2p:ou_abc") {
			t.Errorf("missing routing_key in output: %q", got)
		}
		if !strings.Contains(got, "sender_id: ou_abc") {
			t.Errorf("missing sender_id in output: %q", got)
		}
		if !strings.HasSuffix(got, "hello") {
			t.Errorf("original prompt not preserved: %q", got)
		}
	})

	t.Run("no injection when channel key empty", func(t *testing.T) {
		req := &ExecuteRequest{
			Prompt:    "hello",
			AppConfig: appCfg,
		}
		got := injectRoutingContext(req)
		if got != "hello" {
			t.Errorf("expected unmodified prompt, got: %q", got)
		}
	})

	t.Run("group channel key", func(t *testing.T) {
		req := &ExecuteRequest{
			Prompt:     "ping",
			ChannelKey: "group:oc_xyz:cli_app1",
			SenderID:   "ou_def",
			AppConfig:  appCfg,
		}
		got := injectRoutingContext(req)
		if !strings.Contains(got, "routing_key: group:oc_xyz") {
			t.Errorf("expected group routing_key, got: %q", got)
		}
	})

	t.Run("thread maps to group routing key", func(t *testing.T) {
		req := &ExecuteRequest{
			Prompt:     "ping",
			ChannelKey: "thread:oc_xyz:tid_001:cli_app1",
			SenderID:   "ou_def",
			AppConfig:  appCfg,
		}
		got := injectRoutingContext(req)
		if !strings.Contains(got, "routing_key: group:oc_xyz") {
			t.Errorf("expected thread→group routing_key, got: %q", got)
		}
	})
}
