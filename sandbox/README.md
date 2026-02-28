# Development Sandbox

This directory is for testing the CIO - Chief Intelligence Officer application during development.

## Usage

Run the application from this directory to test initialization and plugin workflows:

```bash
cd sandbox

# Test fresh initialization
../cio init

# Test plugin commands
../cio plugin list
../cio plugin install cio
../cio plugin create my-custom-domain

# Test advisory sessions
../cio ask "Should we migrate to Kubernetes?"
```

## Structure

When you run `cto init`, this directory will have:

```
sandbox/
├── .gitkeep
├── README.md
└── .cio/           # Created by cto init
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
rm -rf .cio
```

## Notes

- This directory is gitignored - local changes won't be committed
- Use this for manual testing during development
- The `.cio/` folder mirrors what users will have in their projects
