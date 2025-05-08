# Flora Plugins for OneMount

This directory contains Flora plugins for use with JetBrains IDEs in the OneMount project.

## What is Flora?

Flora is a plugin system for JetBrains IDEs that allows extending the functionality of the IDE with JavaScript plugins. These plugins can be used to automate tasks, integrate with external tools, and enhance the development workflow.

## Available Plugins

### Open GitHub Issue

A plugin to open tasks in JetBrains IDE by GitHub issue number. This plugin is designed to work with Junie AI in the OneMount project.

#### Installation

1. Ensure you have Flora installed in your JetBrains IDE
2. Copy the `open-github-issue.js` file to your Flora plugins directory
   - For GoLand: `~/.config/JetBrains/GoLand<version>/flora/plugins/`
   - For IntelliJ IDEA: `~/.config/JetBrains/IntelliJIdea<version>/flora/plugins/`
   - For other JetBrains IDEs: `~/.config/JetBrains/<IDE><version>/flora/plugins/`
3. Restart your JetBrains IDE
4. Verify the plugin is loaded by checking the Flora plugin manager

#### Usage

You can use the plugin in two ways:

1. **From the Flora command palette**:
   - Open the Flora command palette (default shortcut: `Alt+Shift+F`)
   - Type `openGitHubIssue` and press Enter
   - Enter the GitHub issue number when prompted

2. **From Junie AI**:
   - When working with Junie AI, you can ask it to open a GitHub issue
   - Basic example: "Open GitHub issue 123"
   - With custom IDE: "Open GitHub issue 123 in IntelliJ IDEA"
   - Junie AI will use the Flora plugin to open the issue as a task in your JetBrains IDE

#### Example

```javascript
// Basic usage - Open GitHub issue #123
Flora.commands.openGitHubIssue(123);

// With custom IDE - Open GitHub issue #123 in IntelliJ IDEA
Flora.commands.openGitHubIssue(123, { ide: "idea" });
```

#### Configuration

The plugin supports the following configuration options:

- **defaultIde**: The default JetBrains IDE to use when auto-detection fails
  - Type: string
  - Default: "goland"
  - Example values: "idea", "pycharm", "webstorm", etc.

You can configure these options in the Flora plugin manager.

## Development

### Creating a New Plugin

To create a new Flora plugin:

1. Create a new JavaScript file in this directory
2. Define the plugin metadata and commands
3. Implement the command functions
4. Register the plugin with Flora
5. Export the plugin

### Plugin Structure

A typical Flora plugin has the following structure:

```javascript
// Define the plugin metadata
const metadata = {
  name: "Plugin Name",
  description: "Plugin description",
  version: "1.0.0",
  author: "Author Name",
  commands: [
    {
      name: "commandName",
      description: "Command description",
      usage: "commandName <arg1> <arg2>"
    }
  ]
};

// Implement command functions
async function commandName(arg1, arg2) {
  // Command implementation
}

// Register the plugin with Flora
function registerPlugin() {
  return {
    metadata: metadata,
    commands: {
      commandName: commandName
    }
  };
}

// Export the plugin
module.exports = registerPlugin();
```

### Testing

To test a Flora plugin:

1. Install the plugin in your JetBrains IDE
2. Open the Flora console (default shortcut: `Alt+Shift+C`)
3. Execute the plugin command with test arguments
4. Check the console output for errors or success messages

## Integration with Junie AI

Flora plugins can be integrated with Junie AI to provide additional functionality. Junie AI can use Flora plugins to perform actions in the JetBrains IDE, such as opening tasks by GitHub issue number.

To integrate a Flora plugin with Junie AI:

1. Ensure the plugin is installed in your JetBrains IDE
2. Configure Junie AI to use the plugin
3. Use natural language to ask Junie AI to perform actions using the plugin

## Troubleshooting

If you encounter issues with Flora plugins:

1. Check the Flora console for error messages
2. Verify that the plugin is installed in the correct directory
3. Restart your JetBrains IDE
4. Check the JetBrains IDE log for any errors related to Flora plugins
