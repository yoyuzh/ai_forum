import { db } from "../api/db";
import { sseEmitter } from "./emitter";
import { AIAgent } from "../api/types";

// Generates an LLM-styled response template based on agent personality rules.
function generateAgentText(agentName: string, title: string, category: string): string {
  const time = new Date().toLocaleTimeString();
  switch (agentName) {
    case "ArchTechLead":
      return `### Design Critique: "${title}"
Reviewing the proposal at ${time}, here are the key architectural observations:
1. **Decoupled Boundaries**: In the context of "${category}", we must establish clear interfaces to prevent tight process-level coupling.
2. **Resource Metrics**: Monolithic structures are easier to profile. Validate database connection pools before scaling hot paths.
3. **Recommendation**: Write integration tests covering state consistency under network partitioned scenarios first.`;

    case "GrowthProductManager":
      return `Wow! The proposal around "${title}" touches some vital user retention paths! 🚀
- Have we defined core product metrics to measure this change?
- From a usability standpoint, let's keep the user loops short and clean.
- Let's schedule a brief sync to refine the telemetry dashboard!`;

    case "DevilsAdvocate":
      return `Let's pause and do a simple sanity check on this "${title}" proposal.
- **Why are we adding this complexity?** It feels like premature optimization.
- Monolithic setups don't need additional microservice dependencies. What happens if the network layer breaks down?
- Keep it simple, test the basic path, and don't introduce fancy solutions for simple problems.`;

    default:
      return `Interesting post on "${title}". From my perspective as an AI collaborator, maintaining clean boundaries is key.`;
  }
}

export function runBackgroundAISimulation(postId: number, commentId: number | null) {
  const post = db.getPost(postId);
  if (!post) return;

  const comments = db.getComments(postId);
  const targetComment = commentId ? comments.find(c => c.id === commentId) : null;
  
  // If triggered by an AI comment, abort to prevent infinite loops
  if (targetComment && targetComment.author.isAi) {
    return;
  }

  const agents = db.getAgents().filter(a => a.active);
  let replyQueue: Array<{ agent: AIAgent; triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP"; willingness: number }> = [];

  if (commentId === null) {
    // Post Auto Reply Flow
    agents.forEach(agent => {
      if (agent.allowAutoReply) {
        const existingAutoReplies = comments.filter(c => c.author.aiAgentId === agent.id && c.parentId === null);
        if (existingAutoReplies.length < agent.maxAutoRepliesPerPost) {
          const willingness = Math.random();
          const decision = willingness >= agent.replyThreshold ? "REPLY" : "IGNORE";
          const reason = decision === "REPLY"
            ? `Willingness score (${willingness.toFixed(2)}) exceeded threshold (${agent.replyThreshold}).`
            : `Willingness score (${willingness.toFixed(2)}) did not satisfy threshold (${agent.replyThreshold}).`;

          const log = db.createDecisionLog({
            postId,
            commentId: null,
            aiAgentId: agent.id,
            aiAgentName: agent.name,
            triggerType: "POST_AUTO",
            willingnessScore: willingness,
            thresholdValue: agent.replyThreshold,
            decision,
            reason
          });
          sseEmitter.emit("decision_log.created", log);

          if (decision === "REPLY") {
            replyQueue.push({ agent, triggerType: "POST_AUTO", willingness });
          }
        }
      }
    });
  } else if (targetComment) {
    // Comment Flow (Mentions or Followups)
    agents.forEach(agent => {
      // 1. Check Mentions
      const mentionRegex = new RegExp(`@${agent.name}\\b`, "i");
      const isMentioned = mentionRegex.test(targetComment.content);

      if (isMentioned && agent.allowMentionReply) {
        // Mention triggers always bias willingness to high
        const willingness = Math.min(1.0, Math.random() + 0.3);
        const decision = willingness >= agent.replyThreshold ? "REPLY" : "IGNORE";
        const reason = decision === "REPLY"
          ? `Agent was explicitly mentioned in comment ${commentId}. Willingness (${willingness.toFixed(2)}) passed threshold.`
          : `Agent was mentioned but willingness (${willingness.toFixed(2)}) did not satisfy threshold (${agent.replyThreshold}).`;

        const log = db.createDecisionLog({
          postId,
          commentId,
          aiAgentId: agent.id,
          aiAgentName: agent.name,
          triggerType: "MENTION",
          willingnessScore: willingness,
          thresholdValue: agent.replyThreshold,
          decision,
          reason
        });
        sseEmitter.emit("decision_log.created", log);

        if (decision === "REPLY") {
          replyQueue.push({ agent, triggerType: "MENTION", willingness });
        }
      } else if (!isMentioned && agent.allowFollowupReply) {
        // 2. Check Followup (User replied to an AI agent's comment)
        if (targetComment.parentId !== null) {
          const parentComment = comments.find(c => c.id === targetComment.parentId);
          if (parentComment && parentComment.author.isAi && parentComment.author.aiAgentId === agent.id) {
            const existingFollowups = comments.filter(c => c.author.aiAgentId === agent.id && c.parentId !== null);
            if (existingFollowups.length < agent.maxFollowupRepliesPerPost) {
              const willingness = Math.random();
              const decision = willingness >= agent.replyThreshold ? "REPLY" : "IGNORE";
              const reason = decision === "REPLY"
                ? `Followup triggered by user response to ${agent.name}'s comment. Willingness (${willingness.toFixed(2)}) passed threshold.`
                : `Followup triggered but willingness (${willingness.toFixed(2)}) did not satisfy threshold (${agent.replyThreshold}).`;

              const log = db.createDecisionLog({
                postId,
                commentId,
                aiAgentId: agent.id,
                aiAgentName: agent.name,
                triggerType: "FOLLOWUP",
                willingnessScore: willingness,
                thresholdValue: agent.replyThreshold,
                decision,
                reason
              });
              sseEmitter.emit("decision_log.created", log);

              if (decision === "REPLY") {
                replyQueue.push({ agent, triggerType: "FOLLOWUP", willingness });
              }
            }
          }
        }
      }
    });
  }

  if (replyQueue.length === 0) {
    // If no agent replies, transition post state to COMPLETED after a brief period
    setTimeout(() => {
      const latestPost = db.getPost(postId);
      if (latestPost) {
        db.updatePost(postId, { aiStatus: "COMPLETED" });
        sseEmitter.emit("post.updated", db.getPost(postId));
      }
    }, 1000);
    return;
  }

  // Mark post processing state
  db.updatePost(postId, { aiStatus: "PROCESSING" });
  sseEmitter.emit("post.updated", db.getPost(postId));

  // Process the reply queue with staggered delays
  replyQueue.forEach((item, index) => {
    const { agent, triggerType } = item;
    
    setTimeout(() => {
      // Create pending task
      const task = db.createTask({
        postId,
        parentCommentId: targetComment ? targetComment.id : null,
        targetCommentId: targetComment ? targetComment.id : null,
        aiAgentId: agent.id,
        triggerType,
        status: "PENDING",
        prompt: `System: ${agent.systemPrompt}\nStyle: ${agent.stylePrompt}\nContext: ${post.title}`,
        result: "",
        errorMessage: "",
        startedAt: null,
        finishedAt: null
      });
      sseEmitter.emit("task.created", task);

      // Transition to PROCESSING
      setTimeout(() => {
        db.updateTask(task.id, {
          status: "PROCESSING",
          startedAt: new Date().toISOString()
        });
        const processingTask = db.getTasks().find(t => t.id === task.id);
        if (processingTask) {
          sseEmitter.emit("task.updated", processingTask);
        }

        // Transition to COMPLETED and generate comment
        setTimeout(() => {
          const replyText = generateAgentText(agent.name, post.title, post.category);
          
          const createdComment = db.createComment({
            postId,
            parentId: targetComment ? targetComment.id : null,
            content: replyText,
            author: {
              username: agent.name,
              avatar: agent.avatar,
              isAi: true,
              aiAgentId: agent.id
            }
          });
          sseEmitter.emit("comment.created", createdComment);

          db.updateTask(task.id, {
            status: "COMPLETED",
            result: replyText,
            finishedAt: new Date().toISOString()
          });
          const completedTask = db.getTasks().find(t => t.id === task.id);
          if (completedTask) {
            sseEmitter.emit("task.updated", completedTask);
          }

          // Update post stats
          const latestPost = db.getPost(postId);
          if (latestPost) {
            const updatedAvatars = Array.from(new Set([...latestPost.aiAvatars, agent.avatar]));
            db.updatePost(postId, {
              aiResponsesCount: latestPost.aiResponsesCount + 1,
              aiAvatars: updatedAvatars
            });
            sseEmitter.emit("post.updated", db.getPost(postId));
          }
        }, 2000); // 2 seconds processing delay

      }, 1500); // 1.5 seconds pending delay

    }, index * 1000); // Stagger task creation by 1 second per agent
  });

  // Schedule final COMPLETED status transition
  const totalDuration = (replyQueue.length * 1000) + 4000;
  setTimeout(() => {
    const latestPost = db.getPost(postId);
    if (latestPost) {
      db.updatePost(postId, { aiStatus: "COMPLETED" });
      sseEmitter.emit("post.updated", db.getPost(postId));
    }
  }, totalDuration);
}
