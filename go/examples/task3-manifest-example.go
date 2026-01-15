// Package main provides a reference example for Task 3 (Synthesis Architecture)
// showing how to use the Buf-generated manifest proto to convert SDK types to AgentManifest.
//
// This file is a REFERENCE ONLY for Task 3 implementation.
// It will be removed once actual synthesis architecture is implemented.
package main

import (
	"time"

	// Import Buf-generated proto packages
	agentv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/agent/v1"
	sdk "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/commons/sdk"
)

// Task3Example shows how to create an AgentManifest proto from SDK configuration.
//
// This demonstrates the conversion logic needed in internal/synth/converter package.
func Task3Example() *agentv1.AgentManifest {
	// Step 1: Create SDK metadata
	metadata := &sdk.SdkMetadata{
		Language:     "go",
		Version:      "0.1.0",
		GeneratedAt:  time.Now().Unix(),
		ProjectName:  "my-project",
	}

	// Step 2: Create AgentBlueprint (the core agent configuration)
	agent := &agentv1.AgentBlueprint{
		Name:         "example-agent",
		Instructions: "You are a helpful coding assistant that follows best practices.",
		Description:  "A code review assistant with security expertise",

		// Skills (converted from SDK skill.Skill types)
		Skills: []*agentv1.ManifestSkill{
			// Platform skill reference
			{
				Id: "skill-1",
				Source: &agentv1.ManifestSkill_Platform{
					Platform: &agentv1.PlatformSkillReference{
						Name: "code-review",
					},
				},
			},
			// Inline skill definition
			{
				Id: "skill-2",
				Source: &agentv1.ManifestSkill_Inline{
					Inline: &agentv1.InlineSkillDefinition{
						Name:            "custom-skill",
						Description:     "A custom skill for specific tasks",
						MarkdownContent: "# Custom Skill\n\nThis skill provides specialized functionality.",
					},
				},
			},
		},

		// MCP Servers (converted from SDK mcpserver.MCPServer types)
		McpServers: []*agentv1.ManifestMcpServer{
			{
				Name:         "github",
				EnabledTools: []string{"create_issue", "list_repos"},
				ServerType: &agentv1.ManifestMcpServer_Stdio{
					Stdio: &agentv1.ManifestStdioServer{
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-github"},
						EnvPlaceholders: map[string]string{
							"GITHUB_TOKEN": "${GITHUB_TOKEN}",
						},
						WorkingDir: "/app",
					},
				},
			},
			{
				Name: "api-service",
				ServerType: &agentv1.ManifestMcpServer_Http{
					Http: &agentv1.ManifestHttpServer{
						Url: "https://mcp.example.com",
						Headers: map[string]string{
							"Authorization": "Bearer ${API_TOKEN}",
						},
						TimeoutSeconds: 60,
					},
				},
			},
		},

		// Sub-agents (converted from SDK subagent.SubAgent types)
		SubAgents: []*agentv1.ManifestSubAgent{
			{
				Source: &agentv1.ManifestSubAgent_Inline{
					Inline: &agentv1.InlineSubAgentDefinition{
						Name:         "security-reviewer",
						Instructions: "Focus on security vulnerabilities and best practices.",
						Description:  "Specialized security review agent",
					},
				},
			},
		},

		// Environment variables (converted from SDK environment.Variable types)
		EnvironmentVariables: []*agentv1.ManifestEnvironmentVariable{
			{
				Name:         "LOG_LEVEL",
				Description:  "Logging level for the agent",
				IsSecret:     false,
				DefaultValue: "info",
				Required:     false,
			},
			{
				Name:        "API_KEY",
				Description: "API key for external service",
				IsSecret:    true,
				Required:    true,
			},
		},
	}

	// Step 3: Create the complete AgentManifest
	manifest := &agentv1.AgentManifest{
		SdkMetadata: metadata,
		Agents:      []*agentv1.AgentBlueprint{agent},
	}

	return manifest
}

// Task3ConversionNotes documents the conversion mapping for Task 3 implementation.
//
// SDK Type → Proto Type Mapping:
//
// 1. agent.Agent → agentv1.AgentBlueprint
//    - Name() → name
//    - Instructions() → instructions
//    - Skills() → skills ([]ManifestSkill)
//    - MCPServers() → mcp_servers ([]ManifestMcpServer)
//    - SubAgents() → sub_agents ([]ManifestSubAgent)
//    - EnvironmentVariables() → environment_variables ([]ManifestEnvironmentVariable)
//
// 2. skill.Skill → agentv1.ManifestSkill
//    - Generate unique ID for each skill
//    - IsPlatformReference() → source = platform (PlatformSkillReference)
//    - IsRepositoryReference() → source = org (OrgSkillReference)
//    - IsInline() → source = inline (InlineSkillDefinition)
//      - Name() → inline.name
//      - Markdown() → inline.markdown_content
//
// 3. mcpserver.MCPServer → agentv1.ManifestMcpServer
//    - Name() → name
//    - EnabledTools() → enabled_tools
//    - Type() == Stdio → server_type = stdio (ManifestStdioServer)
//      - Command() → stdio.command
//      - Args() → stdio.args
//      - Environment() → stdio.env_placeholders (map)
//      - WorkingDir() → stdio.working_dir
//    - Type() == HTTP → server_type = http (ManifestHttpServer)
//      - URL() → http.url
//      - Headers() → http.headers (map)
//      - QueryParams() → http.query_params (map)
//      - TimeoutSeconds() → http.timeout_seconds
//    - Type() == Docker → server_type = docker (ManifestDockerServer)
//      - Image() → docker.image
//      - Args() → docker.args
//      - Environment() → docker.env_placeholders (map)
//      - Volumes() → docker.volumes ([]ManifestVolumeMount)
//      - Ports() → docker.ports ([]ManifestPortMapping)
//      - Network() → docker.network
//      - ContainerName() → docker.container_name
//
// 4. subagent.SubAgent → agentv1.ManifestSubAgent
//    - IsInline() → source = inline (InlineSubAgentDefinition)
//      - Name() → inline.name
//      - Instructions() → inline.instructions
//    - IsReference() → source = reference (ReferencedSubAgent)
//      - AgentInstanceId() → reference.agent_instance_id
//
// 5. environment.Variable → agentv1.ManifestEnvironmentVariable
//    - Key() → name
//    - Value() → default_value
//    - IsSecret() → is_secret
//    - Required() → required
//    - Description() → description
//
// SDK Metadata:
// - Language: "go"
// - Version: SDK version (from build info or constant)
// - GeneratedAt: Unix timestamp
// - ProjectName: From stigmer.json or config
//
// File Output:
// - Serialize manifest to: manifest.pb (binary protobuf)
// - CLI reads manifest.pb and converts to platform API types
