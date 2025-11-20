Feature: src-commands_design-serve-port-conflict

  As a developer
  I want the design serve command to detect and handle port conflicts intelligently
  So that I can manage multiple Structurizr Lite instances without manual intervention

  Rule: Port conflict detection for Structurizr Lite on port 8080

    @ov @depm:src-commands @deps:docker
    Scenario: Detect when port 8080 is already in use
      Given the design serve command is configured to use port 8080
      And another container is already running on port 8080
      When the design serve command attempts to start
      Then the command identifies the conflicting container
      And presents the conflict to the user

    @ov @depm:src-commands @deps:docker
    Scenario: Identify which container is using the port
      Given a container is occupying port 8080
      When the design serve command detects the conflict
      Then the command displays the container name
      And the command displays the container image name
      And the command displays the container process information

  Rule: Interactive conflict resolution with user prompts

    @ov @depm:src-commands @deps:docker
    Scenario: User chooses to stop existing container and start new one
      Given port 8080 is in use by another container
      And the design serve command detects this conflict
      When the user selects the option to stop the existing container
      Then the existing container is stopped
      And the new Structurizr Lite container starts on port 8080
      And the user receives confirmation of success

    @ov @depm:src-commands @deps:docker
    Scenario: User chooses to keep existing container running
      Given port 8080 is in use by another container
      And the design serve command detects this conflict
      When the user selects the option to keep the existing container
      Then the conflicting container continues running
      And the design serve command aborts
      And the user is informed the operation was cancelled

    @ov @depm:src-commands @deps:docker
    Scenario: User chooses to abort the operation
      Given port 8080 is in use by another container
      And the design serve command detects this conflict
      When the user selects the option to abort
      Then the design serve command stops
      And no containers are started or stopped
      And the user receives confirmation of the abort

  Rule: Automatic conflict resolution without prompting

    @ov @depm:src-commands @deps:docker
    Scenario: Automatically stop conflicting container with flag
      Given port 8080 is in use by another container
      And the design serve command is run with the auto-stop flag
      When the command executes
      Then the existing container is automatically stopped
      And the new Structurizr Lite container starts on port 8080
      And no user interaction is required

    @ov @depm:src-commands @deps:docker
    Scenario: Respect explicit auto-stop flag during startup
      Given the user runs design serve with --auto-stop-conflict flag
      And port 8080 is occupied by a running container
      When the design serve command starts
      Then the conflicting container is stopped without prompting
      And Structurizr Lite starts successfully
      And the design workspace is accessible

  Rule: Error handling for port conflicts

    @ov @depm:src-commands @deps:docker
    Scenario: Display clear error message when port is unavailable
      Given port 8080 is in use
      And the user cannot access Docker daemon to identify container
      When the design serve command attempts to start
      Then a clear error message is displayed
      And the error message indicates the port conflict
      And the command provides troubleshooting guidance

    @ov @depm:src-commands @deps:docker
    Scenario: Handle Docker daemon unavailable during conflict detection
      Given Docker daemon is not running
      When the design serve command attempts to detect port conflicts
      Then the command gracefully handles the missing daemon
      And the user is informed Docker is required for conflict detection
      And the command offers to proceed without conflict management