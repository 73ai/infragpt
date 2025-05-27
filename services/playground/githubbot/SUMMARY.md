# GitHub Integration Chatbot - Implementation Overview

## Project Context

This application implements a Slack chatbot that can analyze code from GitHub repositories. When users share GitHub links in Slack, the bot automatically detects them, fetches the code, and provides an analysis. The bot uses OAuth for GitHub authentication and Slack's Socket Mode for event handling, eliminating the need for public endpoints for Slack events.

## Core Functionality

1. **GitHub Link Detection**: Automatically detects when users share GitHub repository or file links in Slack.

2. **User Authentication**: Manages GitHub OAuth flow to get permissions for accessing repositories.

3. **Repository Selection**: Allows users to grant access to all repositories or select specific ones they want the bot to analyze.

4. **Content Fetching**: Retrieves file content, directory structures, or repository information from GitHub based on the link type.

5. **Code Analysis**: Currently implements a mock analysis service that could be replaced with an actual AI service.

## Technical Architecture

### Key Components

1. **TokenStore**: Thread-safe in-memory storage for GitHub OAuth tokens and repository permissions.

2. **Slack Socket Mode Client**: Receives events via WebSockets instead of HTTP webhooks.

3. **GitHub OAuth Integration**: Handles authentication and maintains access tokens.

4. **Link Parser**: Extracts information from various GitHub URL formats using regular expressions.

5. **Content Fetcher**: Uses the GitHub API to retrieve content based on link type.

### Authentication Flow

1. When a user shares a GitHub link without being authenticated, they receive a link to connect their GitHub account.

2. Upon clicking, they're redirected to GitHub's OAuth authorization page.

3. After authorizing, they're taken to a repository selection page where they can choose which repositories to grant access to.

4. The bot stores this information and uses it for future link analysis.

### Event Processing Flow

1. User posts a message containing a GitHub link in Slack.

2. The Socket Mode client receives the event and processes it.

3. The bot extracts GitHub links and checks if the user is authenticated.

4. If authenticated and with proper permissions, the bot fetches the content from GitHub.

5. The bot analyzes the content and posts the results back to the Slack channel.

## Implementation Details

### Slack Integration

- Uses Socket Mode for receiving events via WebSockets
- Requires both a Bot Token (`xoxb-`) and an App-Level Token (`xapp-`)
- Handles message events specifically to detect GitHub links

### GitHub Integration

- Supports multiple link formats (repositories, files, directories)
- Handles permissions at the repository level
- Implements size and format checking to avoid processing large or binary files

### Security Considerations

- Uses OAuth state parameter to prevent CSRF attacks
- Stores tokens in memory (would need to be enhanced for production)
- Implements repository-level permission checking

## Limitations & Future Enhancements

1. **Mock Analysis**: The current implementation uses a simple mock analysis that should be replaced with an actual AI service.

2. **In-Memory Storage**: Tokens are stored in memory and will be lost on server restart. Should be replaced with a database.

3. **Limited Error Handling**: Error handling could be enhanced, particularly for rate limiting and token expiration.

4. **Repository Structure Analysis**: Currently has limited support for analyzing entire repositories or directories.

## Setup Requirements

To run this application, you need:

1. A Slack app with Socket Mode enabled and appropriate scopes (`channels:history`, `chat:write`)
2. A GitHub OAuth app with proper callback URL configuration
3. Go 1.16+ with required dependencies
4. Public URL for OAuth callback endpoints (e.g., using ngrok for development)

## Integration Points

For future contributors looking to extend this application:

1. **AI Service Integration**: Replace the `analyzeCode()` function with calls to a real AI service.

2. **Database Integration**: Add a persistence layer to store tokens and user preferences.

3. **Enhanced Error Recovery**: Implement token refresh and better handling of API rate limits.

4. **Repository Structure Analysis**: Improve the analysis of repository structure with recursive traversal.

5. **UI Improvements**: Enhance the repository selection UI and add more interactive elements.

This overview should provide new contributors with a solid understanding of the application's purpose, architecture, and areas for potential enhancement.