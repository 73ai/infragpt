# GitHub Integration Chatbot for Slack

This is a Go application that provides a Slack chatbot which can analyze GitHub repository code. The bot detects GitHub links in Slack messages, fetches the code, and provides analysis.

## Features

- Web-based GitHub OAuth integration
- Repository-specific permission management
- Automatic GitHub link detection in Slack messages
- Support for file, directory, and repository links
- Mock code analysis (can be replaced with real AI service)
- Error handling for permission issues, large files, and unsupported formats

## Prerequisites

- Go 1.16 or higher
- A Slack app with appropriate permissions
- A GitHub OAuth app
- Publicly accessible URL for callback endpoints (use ngrok for local development)

## Dependencies

- github.com/google/go-github/v39
- github.com/slack-go/slack
- golang.org/x/oauth2

## Setup Instructions

### 1. Create a GitHub OAuth App

1. Go to GitHub > Settings > Developer Settings > OAuth Apps > New OAuth App
2. Set the callback URL to `https://your-domain.com/github/callback`
3. Note the Client ID and Client Secret

### 2. Create a Slack App

1. Go to https://api.slack.com/apps and create a new app
2. Enable Event Subscriptions and subscribe to `message.channels` events
3. Set the Request URL to `https://your-domain.com/slack/events`
4. Create a Bot User and install the app to your workspace
5. Note the Bot User OAuth Token and Signing Secret

### 3. Configure the Application

Edit the `main()` function to use your actual credentials:

```go
config := Config{
    GithubClientID:     "your-github-client-id",
    GithubClientSecret: "your-github-client-secret",
    SlackToken:         "your-slack-token",
    SlackSigningSecret: "your-slack-signing-secret",
    BaseURL:            "https://your-domain.com",
}
```

### 4. Build and Run

```bash
go build -o github-chatbot
./github-chatbot
```

The server will start on port 8080.

## Application Flow

1. A user posts a GitHub link in a Slack channel
2. The bot detects the link and checks if the user is authenticated
3. If not authenticated, the bot sends an authentication link
4. The user authorizes the app and selects which repositories to grant access to
5. Once authenticated, the bot fetches the content and analyzes it
6. The analysis is posted back to the Slack channel

## Extending the Application

### Replacing the Mock Analysis

The current implementation uses a mock analysis function. To integrate with a real AI service:

1. Modify the `analyzeCode()` function in the main file
2. Add appropriate API clients and request handling
3. Format the AI response appropriately for Slack

### Improving Error Handling

The current implementation has basic error handling. You might want to:

1. Add better handling for GitHub API rate limits
2. Implement token refresh when tokens expire
3. Add more detailed error messages for users

### Persistence

The current implementation uses an in-memory `sync.Map`. For production:

1. Implement a database layer for token storage
2. Add proper encryption for stored tokens
3. Implement session management for the OAuth flow

## Security Considerations

- Ensure your GitHub OAuth app is properly secured
- Store tokens securely (use encryption in production)
- Validate all Slack request signatures
- Implement proper CSRF protection for the OAuth flow

## License

[Your License Here]