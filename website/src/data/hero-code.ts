export const heroCode = `# Validate all Go code blocks in your docs
md-go-validator .

# JSON output for CI
md-go-validator -f json -o results.json .

# Multiple languages
md-go-validator -l go,typescript,rust .`;
