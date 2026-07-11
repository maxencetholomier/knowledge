# Kl (Knowledge) - Zettelkasten CLI Tool

Knowledge is a command-line tool for managing [markdown-based notes](./docs/plain-text-importance.md) in a [Zettelkasten](./docs/zettlekasten.md)-style system.

Compared to other [related projects](#related-projects), Knowledge offers :

- Note availability on smartphones (read and write access) with [Joplin](https://github.com/laurent22/joplin)
- Proper version control on computer
- Spaced repetition learning through [Anki](https://github.com/ankitects/anki)

For naming and file organization, see the [Conventions Guide](./docs/conventions.md).

## System Architecture

See the [Architecture Guide](./docs/architecture.md) for how Knowledge integrates with Joplin and Anki.

## Dependencies

### Required Tools

- **[Rg](https://github.com/BurntSushi/ripgrep)**: For fast grep search

### Optional Tools

- **[Anki](https://apps.ankiweb.net/)**: For spaced repetition learning (`anki export` command)
- **[Flameshot](https://github.com/flameshot-org/flameshot)**: For screenshots (`image` command)
- **[Fzf](https://github.com/junegunn/fzf)**: For interactive fuzzy finding
- **[Inkscape](https://inkscape.org/)**: For diagrams (`schema` command)
- **[Joplin](https://github.com/laurent22/joplin)**: For cross-platform synchronization

## Installation

See the [Installation Guide](./docs/installation.md) for install, update, shell completion, man pages, and Joplin setup instructions.

## Usage

For a quick, hands-on walkthrough, see the [Comprehensive Example](./docs/comprehensive_exemple.md).

For detailed usage examples and workflows, see [Usage Guide](./docs/usage.md).

## Roadmap

See the [Roadmap](./docs/roadmap.md) for planned features and improvements.

## Related Projects

- [Zk](https://github.com/zk-org/zk)
- [Neuron](https://github.com/srid/neuron)
- [Emanote](https://emanote.srid.ca/)
- [Sirupsen's zk](https://github.com/sirupsen/zk)
- [Zk-spaced](https://github.com/matze/zk-spaced)
