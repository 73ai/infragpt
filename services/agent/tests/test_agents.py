"""Tests for the agent system."""

import pytest

from src.agents import AgentSystem, ConversationAgent, RCAAgent, MainAgent
from src.models.agent import AgentRequest
from src.models.context import AgentContext


class TestConversationAgent:
    """Tests for the conversation agent."""

    @pytest.fixture
    def agent(self):
        return ConversationAgent()

    @pytest.fixture
    def context(self):
        return AgentContext(
            conversation_id="test-123",
            current_message="Hello there!",
            message_history=[],
        )

    @pytest.mark.asyncio
    async def test_can_handle_greeting(self, agent, context):
        """Test that conversation agent handles greetings."""
        context.current_message = "Hello!"
        assert await agent.can_handle(context) is True

    @pytest.mark.asyncio
    async def test_can_handle_question(self, agent, context):
        """Test that conversation agent handles general questions."""
        context.current_message = "What can you do?"
        assert await agent.can_handle(context) is True

    @pytest.mark.asyncio
    async def test_does_not_handle_errors(self, agent, context):
        """Test that conversation agent does not handle error messages."""
        context.current_message = "The server is down and throwing 500 errors"
        assert await agent.can_handle(context) is False

    @pytest.mark.asyncio
    async def test_process_greeting(self, agent, context):
        """Test processing a greeting message."""
        context.current_message = "Hello!"
        response = await agent.process(context)

        assert response.success is True
        assert response.agent_type == "conversation"
        assert "Hello" in response.response_text
        assert response.confidence > 0


class TestRCAAgent:
    """Tests for the RCA agent."""

    @pytest.fixture
    def agent(self):
        return RCAAgent()

    @pytest.fixture
    def context(self):
        return AgentContext(
            conversation_id="test-123",
            current_message="The server is down",
            message_history=[],
        )

    @pytest.mark.asyncio
    async def test_can_handle_error(self, agent, context):
        """Test that RCA agent handles error messages."""
        context.current_message = "The server is throwing 500 errors"
        assert await agent.can_handle(context) is True

    @pytest.mark.asyncio
    async def test_can_handle_failure(self, agent, context):
        """Test that RCA agent handles failure messages."""
        context.current_message = "The deployment failed"
        assert await agent.can_handle(context) is True

    @pytest.mark.asyncio
    async def test_does_not_handle_greeting(self, agent, context):
        """Test that RCA agent does not handle greetings."""
        context.current_message = "Hello there!"
        assert await agent.can_handle(context) is False

    @pytest.mark.asyncio
    async def test_process_error(self, agent, context):
        """Test processing an error message."""
        context.current_message = "The server is returning 500 errors"
        response = await agent.process(context)

        assert response.success is True
        assert response.agent_type == "rca"
        assert "## Issue Analysis" in response.response_text
        assert response.confidence > 0


class TestMainAgent:
    """Tests for the main orchestrator agent."""

    @pytest.fixture
    def agent(self):
        return MainAgent()

    @pytest.fixture
    def context(self):
        return AgentContext(
            conversation_id="test-123", current_message="Hello!", message_history=[]
        )

    @pytest.mark.asyncio
    async def test_can_always_handle(self, agent, context):
        """Test that main agent can handle any request."""
        assert await agent.can_handle(context) is True

    @pytest.mark.asyncio
    async def test_routes_to_conversation(self, agent, context):
        """Test routing to conversation agent."""
        context.current_message = "Hello there!"
        response = await agent.process(context)

        assert response.success is True
        assert response.agent_type == "conversation"

    @pytest.mark.asyncio
    async def test_routes_to_rca(self, agent, context):
        """Test routing to RCA agent."""
        context.current_message = "The server is down with 500 errors"
        response = await agent.process(context)

        assert response.success is True
        assert response.agent_type == "rca"


class TestAgentSystem:
    """Tests for the agent system."""

    @pytest.fixture
    def system(self):
        return AgentSystem()

    @pytest.fixture
    def agent_request(self):
        return AgentRequest(
            conversation_id="test-123", current_message="Hello!", past_messages=[]
        )

    @pytest.mark.asyncio
    async def test_system_initialization(self):
        """Test that the agent system initializes correctly."""
        system = AgentSystem()
        assert not system.is_ready()

        await system.initialize()
        assert system.is_ready()

    @pytest.mark.asyncio
    async def test_process_conversation_request(self, system, agent_request):
        """Test processing a conversational request."""
        await system.initialize()
        agent_request.current_message = "Hello there!"
        response = await system.process_request(agent_request)

        assert response.success is True
        assert response.agent_type == "conversation"

    @pytest.mark.asyncio
    async def test_process_rca_request(self, system, agent_request):
        """Test processing an RCA request."""
        await system.initialize()
        agent_request.current_message = "The server is returning 500 errors"
        response = await system.process_request(agent_request)

        assert response.success is True
        assert response.agent_type == "rca"

    @pytest.mark.asyncio
    async def test_get_system_status(self, system):
        """Test getting system status."""
        await system.initialize()
        status = system.get_system_status()

        assert status["status"] == "ready"
        assert status["initialized"] is True
        assert status["main_agent_available"] is True
        assert "agents" in status
