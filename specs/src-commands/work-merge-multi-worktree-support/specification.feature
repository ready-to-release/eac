@skip:wip
Feature: src-commands_work-merge-multi-worktree-support

  As a developer
  I want to merge feature branches to main across multiple worktrees
  So that I can manage code changes safely without switching between worktrees manually

  Rule: Main is checked out in another worktree

    @L2 @ov
    Scenario: Merge detects main in another worktree and navigates there
      Given a repository has multiple worktrees
      And the main branch is checked out in worktree-2
      And I am working in worktree-1 with a feature branch checked out
      When I execute the merge command to merge my feature branch to main
      Then the system detects main is checked out in worktree-2
      And the system navigates to worktree-2
      And the merge is performed in worktree-2
      And the operation completes successfully

    @L2 @ov
    Scenario: User sees confirmation of worktree navigation during merge
      Given a repository has multiple worktrees
      And the main branch is checked out in worktree-2
      And I am in worktree-1 with a feature branch
      When I execute the merge command
      Then a message indicates that main is checked out in worktree-2
      And a message confirms the merge will proceed in worktree-2
      And the current worktree remains unchanged after the operation

    @L2 @ov
    Scenario: Merge fails safely when main is checked out elsewhere and merge conflicts
      Given a repository has multiple worktrees
      And the main branch is checked out in worktree-2
      And the feature branch has conflicts with main
      When I execute the merge command from worktree-1
      Then the system navigates to worktree-2
      And the merge attempts to proceed
      And merge conflict indicators are displayed
      And the current worktree remains unchanged

  Rule: Main is not currently checked out in any worktree

    @L2 @ov
    Scenario: Merge checks out main in current worktree when not checked out elsewhere
      Given a repository has multiple worktrees
      And the main branch is not checked out in any worktree
      And I am working in worktree-1 with a feature branch checked out
      When I execute the merge command to merge my feature branch to main
      Then the system checks out main in the current worktree
      And the merge is performed in the current worktree
      And the operation completes successfully

    @L2 @ov
    Scenario: Merge restores original branch after completing merge
      Given a repository has multiple worktrees
      And the main branch is not checked out anywhere
      And I am in worktree-1 on a feature branch
      When I execute the merge command
      Then the system checks out main temporarily
      And the merge completes successfully
      And the original feature branch is restored in the current worktree

    @L2 @ov
    Scenario: User is informed when merge checks out main locally
      Given a repository has multiple worktrees
      And the main branch is not checked out in any worktree
      And I am in worktree-1 on a feature branch
      When I execute the merge command
      Then a message indicates main is not checked out elsewhere
      And a message confirms main will be checked out locally
      And the merge proceeds in the current worktree

  Rule: Multi-worktree detection works correctly

    @L2 @ov
    Scenario: System correctly identifies all worktrees in repository
      Given a repository has three worktrees
      When I execute the merge command
      Then the system identifies all active worktrees
      And the system correctly determines which branch is checked out in each

    @L2 @ov
    Scenario: Merge handles missing worktree gracefully
      Given a repository previously had a worktree that was removed
      And the main branch was last seen in the removed worktree
      And I am in the current worktree on a feature branch
      When I execute the merge command
      Then the system detects the worktree is no longer available
      And the system falls back to checking out main in the current worktree
      And the merge proceeds successfully