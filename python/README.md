# Stigmer Agent SDK - Python

## Status: 🚧 Coming Soon

The Python SDK is currently available in the [Stigmer monorepo](https://github.com/leftbin/stigmer/tree/main/sdk/python) while it undergoes active development.

## Current Location

**Repository**: [github.com/leftbin/stigmer](https://github.com/leftbin/stigmer)  
**Path**: `sdk/python/`  
**Package**: `stigmer` (published to PyPI)

## Installation

```bash
pip install stigmer
```

Or with Poetry:

```bash
poetry add stigmer
```

## Documentation

- **Python SDK Documentation**: [stigmer/sdk/python/README.md](https://github.com/leftbin/stigmer/blob/main/sdk/python/README.md)
- **API Reference**: [docs.stigmer.ai/sdk/python](https://docs.stigmer.ai/sdk/python)
- **Examples**: [stigmer/sdk/python/examples/](https://github.com/leftbin/stigmer/tree/main/sdk/python/examples)

## Quick Example

```python
from stigmer import Agent, Skill, MCPServer

# Create agent with platform skill
agent = Agent(
    name="code-reviewer",
    instructions="Review code for best practices and security",
    skills=[
        Skill.platform("coding-standards"),
        Skill.inline(
            name="security-guidelines",
            markdown_file="skills/security.md"
        ),
    ],
    mcp_servers=[
        MCPServer.stdio(
            name="github",
            command="npx",
            args=["-y", "@modelcontextprotocol/server-github"],
            env_placeholders={"GITHUB_TOKEN": "${GITHUB_TOKEN}"},
        ),
    ],
)
```

## Migration Plan

The Python SDK will be migrated to this repository (`github.com/leftbin/stigmer-sdk`) in a future release:

### Phase 1 (Current)
- ✅ Python SDK lives in monorepo
- ✅ Published to PyPI as `stigmer`
- ✅ Active development and iteration

### Phase 2 (Future)
- 📋 Move to `github.com/leftbin/stigmer-sdk/python`
- 📋 Update package structure
- 📋 Maintain PyPI publishing
- 📋 Update import paths (if needed)
- 📋 Comprehensive migration guide

## Why Keep in Monorepo?

The Python SDK remains in the monorepo temporarily because:

1. **Rapid Iteration**: Easier to iterate on SDK and platform APIs together
2. **Proto Synchronization**: Direct access to proto definitions
3. **Integration Testing**: Easier end-to-end testing with platform
4. **Stability First**: Want Go SDK to establish patterns before Python migration

## Feature Parity

The Python SDK has feature parity with the Go SDK:

- ✅ Agent configuration
- ✅ Platform, organization, and inline skills
- ✅ MCP servers (stdio, HTTP, Docker)
- ✅ Sub-agents (inline and referenced)
- ✅ Environment variables (secrets and configs)
- ✅ File-based content loading
- ✅ Proto-agnostic architecture (planned)

## Contributing

To contribute to the Python SDK:

1. Clone the main repository: `git clone https://github.com/leftbin/stigmer.git`
2. Navigate to Python SDK: `cd sdk/python`
3. Install dependencies: `poetry install`
4. Run tests: `poetry run pytest`
5. See [CONTRIBUTING.md](https://github.com/leftbin/stigmer/blob/main/CONTRIBUTING.md)

## Support

- **Issues**: [Stigmer Issues](https://github.com/leftbin/stigmer/issues) (use label: `sdk:python`)
- **Discussions**: [Stigmer Discussions](https://github.com/leftbin/stigmer/discussions)
- **Documentation**: [docs.stigmer.ai](https://docs.stigmer.ai)

## Timeline

**Expected Migration**: Q2 2026

We'll announce the migration timeline once the Go SDK patterns are established and stable.

---

**For now, please use the Python SDK from the main Stigmer repository.**

[→ Go to Python SDK Documentation](https://github.com/leftbin/stigmer/tree/main/sdk/python)
