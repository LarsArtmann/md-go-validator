# Multi-Language Validation Examples

This document demonstrates the multi-language validation capabilities of md-go-validator.

## Go

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

```go
// Partial code - valid via function wrapping
type User struct {
    Name  string
    Email string
}
```

## TypeScript

```typescript
// Valid TypeScript
interface User {
    name: string;
    email: string;
}

function greet(user: User): string {
    return `Hello, ${user.name}!`;
}
```

```typescript
// TypeScript with generics
class Container<T> {
    private value: T;

    constructor(value: T) {
        this.value = value;
    }

    getValue(): T {
        return this.value;
    }
}
```

## Rust

```rust
// Valid Rust
fn main() {
    println!("Hello, World!");
}
```

```rust
// Struct definition
struct User {
    name: String,
    email: String,
}

impl User {
    fn new(name: &str, email: &str) -> Self {
        User {
            name: name.to_string(),
            email: email.to_string(),
        }
    }
}
```

## Nix

```nix
# Valid Nix expression
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
  ];
}
```

```nix
# Nix function
{ stdenv, lib, fetchFromGitHub }:

stdenv.mkDerivation rec {
  pname = "example";
  version = "1.0.0";

  src = fetchFromGitHub {
    owner = "example";
    repo = "example";
    rev = "v${version}";
    sha256 = lib.fakeSha256;
  };
}
```

## HCL / Terraform

```hcl
# Valid Terraform configuration
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "ExampleInstance"
  }
}
```

## Templ

```templ
// Valid Templ component
package components

type User struct {
    Name  string
    Email string
}

templ UserCard(user User) {
    <div class="user-card">
        <h2>{ user.Name }</h2>
        <p>{ user.Email }</p>
    </div>
}
```

```templ
// Templ page layout
package pages

templ Layout(title string) {
    <!DOCTYPE html>
    <html lang="en">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ title }</title>
        </head>
        <body>
            { children... }
        </body>
    </html>
}
```

## Skipped Blocks

<!-- skip-validate -->

```go
// This code is intentionally invalid
// and will be skipped during validation
func broken( {
    return nil
}
```

```typescript
// Also skipped
const x: InvalidType = 123;
```

## Validation Commands

Validate this file with different language combinations:

```bash
# Validate only Go (default)
md-go-validator EXAMPLES.md

# Validate Go and TypeScript
md-go-validator -l go,typescript EXAMPLES.md

# Validate all supported languages
md-go-validator -l go,templ,typescript,nix,rust,hcl EXAMPLES.md

# Validate with verbose output
md-go-validator -v -l go,typescript EXAMPLES.md

# Output as JSON
md-go-validator -f json -l go,rust EXAMPLES.md
```
