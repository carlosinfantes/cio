# Development Sandbox

This directory is for testing the CTO Advisory Board application during development.

## Usage

Run the application from this directory to test initialization and plugin workflows:

```bash
cd sandbox

# Test fresh initialization
../cto-advisory init

# Test plugin commands
../cto-advisory plugin list
../cto-advisory plugin install cto-advisory
../cto-advisory plugin create my-custom-domain

# Test advisory sessions
../cto-advisory ask "Should we migrate to Kubernetes?"
```

## Structure

When you run `cto init`, this directory will have:

```
sandbox/
├── .gitkeep
├── README.md
└── .cto-advisory/           # Created by cto init
    ├── config.yaml
    ├── context/
    ├── decisions/
    ├── discovery/
    └── plugins/
        ├── installed/       # Downloaded from registry
        └── custom/          # User-created plugins
```

## Cleaning Up

To reset and test fresh initialization:

```bash
rm -rf .cto-advisory
```

## Notes

- This directory is gitignored - local changes won't be committed
- Use this for manual testing during development
- The `.cto-advisory/` folder mirrors what users will have in their projects
