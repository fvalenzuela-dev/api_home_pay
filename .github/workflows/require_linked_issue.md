name: Require Linked Issue

on:
  pull_request:
    types: [opened, edited, reopened, synchronize]
    branches:
      - develop

jobs:
  check-linked-issue:
    runs-on: ubuntu-latest
    if: |
      startsWith(github.head_ref, 'fix/') ||
      startsWith(github.head_ref, 'feat/') ||
      startsWith(github.head_ref, 'feature/') ||
      startsWith(github.head_ref, 'chore/') ||
      startsWith(github.head_ref, 'docs/') ||
      startsWith(github.head_ref, 'style/') ||
      startsWith(github.head_ref, 'refactor/') ||
      startsWith(github.head_ref, 'perf/') ||
      startsWith(github.head_ref, 'test/') ||
      startsWith(github.head_ref, 'build/') ||
      startsWith(github.head_ref, 'ci/') ||
      startsWith(github.head_ref, 'revert/')
    permissions:
      pull-requests: read

    steps:
      - name: Verificar issue vinculada
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          PR_NUMBER: ${{ github.event.pull_request.number }}
          REPO: ${{ github.repository }}
        run: |
          LINKED=$(gh pr view $PR_NUMBER \
            --repo $REPO \
            --json closingIssuesReferences \
            --jq '.closingIssuesReferences | length')

          echo "Issues vinculadas: $LINKED"

          if [ "$LINKED" -eq 0 ]; then
            echo "❌ Este PR no tiene ninguna issue vinculada."
            echo "Vincula una issue en la descripción con:"
            echo "  Closes #123  |  Fixes #123  |  Resolves #123"
            exit 1
          fi

          echo "✅ Issue vinculada correctamente."
