import agent1001 from "../assets/ai-agents/1001.png";
import agent1002 from "../assets/ai-agents/1002.png";
import agent1003 from "../assets/ai-agents/1003.png";
import agent1004 from "../assets/ai-agents/1004.png";
import agent1005 from "../assets/ai-agents/1005.png";
import agent1006 from "../assets/ai-agents/1006.png";
import agent1007 from "../assets/ai-agents/1007.png";
import agent1008 from "../assets/ai-agents/1008.png";
import agent1009 from "../assets/ai-agents/1009.png";
import agent1010 from "../assets/ai-agents/1010.png";
import agent1011 from "../assets/ai-agents/1011.png";
import agent1012 from "../assets/ai-agents/1012.png";

const AI_AGENT_AVATARS: Record<number, string> = {
  1001: agent1001,
  1002: agent1002,
  1003: agent1003,
  1004: agent1004,
  1005: agent1005,
  1006: agent1006,
  1007: agent1007,
  1008: agent1008,
  1009: agent1009,
  1010: agent1010,
  1011: agent1011,
  1012: agent1012,
};

export function aiAgentAvatar(id: number): string | undefined {
  return AI_AGENT_AVATARS[id];
}
