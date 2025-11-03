# TFE AI Prompts

Common AI prompts and commands for working with TFE. Copy-paste these into Claude Code, OpenCode, Gemini, or other AI assistants.

## 📂 Code Analysis Prompts

### Analyze Project Structure
```
Analyze the project structure in this directory. Identify:
1. Main components and their responsibilities
2. Key entry points and configuration files
3. Dependencies and tech stack
4. Potential areas for improvement
```

### Review Architecture
```
Review the architecture of this codebase:
1. Design patterns being used
2. Code organization and modularity
3. Adherence to best practices
4. Scalability concerns
5. Suggest improvements
```

### Security Audit
```
Perform a security audit of this codebase:
1. Check for common vulnerabilities (SQL injection, XSS, etc.)
2. Review authentication/authorization
3. Identify sensitive data handling
4. Check dependency vulnerabilities
5. Suggest security improvements
```

## 🐛 Debugging Prompts

### Debug This Error
```
I'm getting this error:
[PASTE ERROR MESSAGE]

In file: [PASTE FILE PATH from TFE]

Help me:
1. Understand what's causing it
2. Suggest fixes
3. Prevent it in the future
```

### Performance Issue
```
This code is slow:
[PASTE CODE]

Located in: [FILE PATH]

Analyze:
1. Performance bottlenecks
2. Time complexity
3. Memory usage
4. Optimization strategies
```

### Test This Function
```
Write comprehensive tests for this function:
[PASTE FUNCTION]

Include:
1. Unit tests (happy path)
2. Edge cases
3. Error handling
4. Mocking external dependencies
```

## 📝 Documentation Prompts

### Generate README
```
Create a comprehensive README.md for this project at:
[PASTE PROJECT PATH]

Include:
1. Project description
2. Features
3. Installation steps
4. Usage examples
5. Configuration options
6. Contributing guidelines
```

### Document This Code
```
Add comprehensive documentation to this code:
[PASTE CODE]

Include:
1. Function/class docstrings
2. Inline comments for complex logic
3. Usage examples
4. Parameter descriptions
5. Return value documentation
```

### API Documentation
```
Generate API documentation for these endpoints:
[PASTE API FILE]

Format:
1. Endpoint URL and method
2. Request parameters
3. Request body schema
4. Response format
5. Example requests/responses
6. Error codes
```

## 🔧 Refactoring Prompts

### Refactor This Code
```
Refactor this code for better maintainability:
[PASTE CODE]

Focus on:
1. Single Responsibility Principle
2. DRY (Don't Repeat Yourself)
3. Naming clarity
4. Function size
5. Code complexity
```

### Extract Component
```
Extract a reusable component from this code:
[PASTE CODE]

Requirements:
1. Make it generic/configurable
2. Maintain existing functionality
3. Add proper documentation
4. Include usage examples
```

### Modernize Legacy Code
```
Modernize this legacy code:
[PASTE CODE]

Update to:
1. Modern language features
2. Current best practices
3. Better error handling
4. Improved type safety
```

## 🚀 Feature Development Prompts

### Implement Feature
```
Implement this feature:
[DESCRIBE FEATURE]

In project: [PROJECT PATH]

Requirements:
1. Maintain existing code style
2. Add error handling
3. Write tests
4. Update documentation
```

### Design API Endpoint
```
Design a REST API endpoint for:
[DESCRIBE FUNCTIONALITY]

Include:
1. Endpoint structure
2. Request/response schemas
3. Validation rules
4. Error handling
5. Implementation code
```

### Add Configuration
```
Add configuration support for:
[DESCRIBE CONFIG OPTIONS]

Using: [JSON/YAML/TOML/ENV]

Include:
1. Schema definition
2. Default values
3. Validation
4. Loading mechanism
5. Documentation
```

## 🗄️ Database Prompts

### Design Database Schema
```
Design a database schema for:
[DESCRIBE DATA MODEL]

Include:
1. Table definitions
2. Relationships
3. Indexes
4. Constraints
5. Migration scripts
```

### Optimize Query
```
Optimize this database query:
[PASTE QUERY]

Current performance: [DESCRIBE ISSUE]

Provide:
1. Optimized query
2. Explanation of changes
3. Index recommendations
4. Expected performance improvement
```

## 🧪 Testing Prompts

### Write Unit Tests
```
Write unit tests for:
[PASTE CODE]

Using: [TEST FRAMEWORK]

Cover:
1. All public methods
2. Edge cases
3. Error conditions
4. Mocking external dependencies
```

### Create Integration Tests
```
Create integration tests for:
[DESCRIBE FEATURE/API]

Include:
1. Setup/teardown
2. Happy path scenarios
3. Error scenarios
4. Data validation
```

### Test Coverage Analysis
```
Analyze test coverage for:
[PROJECT/FILE PATH]

Identify:
1. Untested code paths
2. Missing edge cases
3. Priority areas for testing
4. Test improvement suggestions
```

## 🛠️ Build & Deployment Prompts

### Create Dockerfile
```
Create a Dockerfile for this project:
[PROJECT PATH]

Requirements:
1. Multi-stage build
2. Minimal image size
3. Security best practices
4. Development and production variants
```

### Setup CI/CD
```
Create GitHub Actions CI/CD pipeline for:
[PROJECT PATH]

Include:
1. Automated testing
2. Linting and formatting
3. Build process
4. Deployment to [PLATFORM]
5. Environment secrets
```

### Environment Configuration
```
Create environment configuration for:
[PROJECT NAME]

Environments: Development, Staging, Production

Include:
1. .env templates
2. Variable documentation
3. Secure defaults
4. Validation rules
```

## 🎨 UI/UX Prompts (for TUI)

### Design TUI Layout
```
Design a TUI layout for:
[DESCRIBE FEATURE]

Using Bubble Tea framework

Include:
1. Component structure
2. Navigation flow
3. Keyboard shortcuts
4. Visual design (lipgloss styles)
```

### Improve UX
```
Improve the user experience of:
[DESCRIBE CURRENT UX]

Suggestions for:
1. Keyboard shortcuts
2. Visual feedback
3. Error messages
4. Help text
5. Workflow optimization
```

## 🔍 Code Review Prompts

### Review Pull Request
```
Review this code change:
[PASTE DIFF or FILE]

Check for:
1. Code quality
2. Potential bugs
3. Performance issues
4. Security concerns
5. Documentation completeness
```

### Suggest Improvements
```
Suggest improvements for:
[PASTE CODE]

Focus on:
1. Readability
2. Maintainability
3. Performance
4. Security
5. Best practices
```

## 🧩 Integration Prompts

### Integrate API
```
Integrate this API into the project:
[API DOCUMENTATION or URL]

Requirements:
1. Type-safe client
2. Error handling
3. Rate limiting
4. Caching strategy
5. Usage examples
```

### Add Third-Party Library
```
Integrate [LIBRARY NAME] into:
[PROJECT PATH]

Include:
1. Installation steps
2. Configuration
3. Usage examples
4. Best practices
5. Migration guide (if replacing existing)
```

## 📊 Analysis Prompts

### Performance Profiling
```
Profile and optimize:
[FILE/FUNCTION PATH]

Analyze:
1. CPU usage
2. Memory allocation
3. I/O operations
4. Bottlenecks
5. Optimization opportunities
```

### Dependency Analysis
```
Analyze dependencies in:
[PROJECT PATH]

Check for:
1. Outdated packages
2. Security vulnerabilities
3. Unused dependencies
4. Circular dependencies
5. Bundle size impact
```

## 🎯 Quick Commands

### Explain Code
```
Explain what this code does:
[PASTE CODE]
```

### Fix Bug
```
Fix this bug:
[PASTE CODE and ERROR]
```

### Add Comments
```
Add helpful comments to:
[PASTE CODE]
```

### Generate Types
```
Generate TypeScript types for:
[PASTE API RESPONSE/DATA]
```

### Convert Format
```
Convert this [FROM FORMAT] to [TO FORMAT]:
[PASTE DATA]
```

## 🔗 TFE-Specific Workflows

### File Path Context
```
I'm in TFE at path: [PASTE PATH from TFE status bar]

Looking at file: [FILENAME]

[YOUR QUESTION OR REQUEST]
```

### Multi-File Context
```
Working with these files (from TFE):
1. [FILE 1 PATH]
2. [FILE 2 PATH]
3. [FILE 3 PATH]

[TASK: refactor/analyze/etc]
```

### Project Context
```
TFE detected this project structure:
- .git/
- package.json
- src/
- tests/

[YOUR QUESTION]
```

---

## 💡 Pro Tips

### Copy File Path from TFE
1. Select file in TFE
2. Press `y` to yank path to clipboard
3. Paste into AI prompt

### Use Prompts Library (F11)
1. Save common prompts in TFE prompts library
2. Add `{{file}}` or `{{path}}` placeholders
3. Use F3 to pick files when running prompt
4. Ctrl+C to copy filled prompt

### Multi-Step Workflow
```bash
# 1. Navigate to file in TFE
# 2. Press 'y' to copy path
# 3. Open Claude Code
# 4. Use prompt template with path
# 5. Get AI assistance
# 6. Press 'e' in TFE to edit file
# 7. Apply AI suggestions
```

---

**Last Updated**: 2024-11-02
**Compatible With**: Claude Code, OpenCode, Gemini, Codex, any AI assistant
