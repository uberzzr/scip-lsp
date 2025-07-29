# protocol package

This directory contains selected content from the gopls/internal/lsp/protocol package, which we can use to supplement the go.lsp.dev/protocol package.  We do not import the whole package as it is internal to gopls and contains significant gopls-specific logic.

However, as it contains various utility functions and mappers that are more generic, the code can be borrowed and added here as needed, in some cases with minor revision to remove dependency on other gopls internal packages.

Issue proposing changes needed to make the gopls version external: https://github.com/golang/go/issues/61338
