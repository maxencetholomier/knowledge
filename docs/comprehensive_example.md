# Comprehensive Example

This walkthrough follows a developer through two days of real usage. It covers:

1. Searching and editing notes (`kl find`, `kl edit`)
2. Creating notes (`kl new`)
3. Synchronizing with Joplin (`kl joplin export`, `import`, `merge`, `clean`)
4. Cleaning up and versioning with git (`kl clean`)
5. Exporting to Anki for spaced repetition (`kl anki export`)

It assumes `kl` is [installed](./installation.md) and your notes directory is [configured](./environnement_configuration_file.md) via the `$K_DIR` environment variable.

## The scenario

You've just gotten a new job as a React developer.

Despite your 10 years of experience, it's been a while since you've worked with JavaScript, and you always forget syntax subtleties, like what is `falsy` or how `array comparison` works.

Going to the internet for each recurring question wastes time. And time is money...

## Day 1 — At the office

You take your coffee and launch a terminal.

Out of curiosity, you check how many notes you already have:

```bash
$ kl list | wc -l
3477
```

Ten years of experience accumulates significant information.

Do you already have notes about the things you keep forgetting?

```bash
$ kl find javascript
1 JavaScript Array Methods
2 JavaScript falsy values
3 ES6 Features Overview
...
```

A note on `falsy` values exists, but nothing on array comparison.

You open the `falsy` note using its number from the result above:

```bash
$ kl edit 2
```

You realize you missed the `""` value, so you update the note and quit your editor.

Then you review `array comparison` on the internet, again... This time you capture it in a new note:

```bash
$ kl new
```

This creates `20240315100000.md`, opens your editor, and you add:

```markdown
# JavaScript Array Comparison

- Arrays are objects, so `==` and `===` compare references, not content
- Use helper functions or loops to compare array content
- JSON.stringify() for simple arrays (has limitations)
```

Your knowledge is now solid. You can come back when you forget this notion again, probably next week...

## Day 1 — Commute home

On the bus, going on Instagram would be a waste of time. It would be nicer to review the notes you added today, on your phone.

This is where [Joplin comes in](./architecture.md): notes exported to Joplin Desktop are synchronized to Joplin Cloud, which Joplin Mobile picks up on your phone.

So before leaving the office, you push your new local notes to Joplin Desktop:

```bash
$ kl joplin export
```

You are ready to go home. During the ride, you review today's notes on your phone, notice a lot has changed in JavaScript during those 10 years, and create and update several notes directly in Joplin Mobile.

## Day 2 — Back at the office

Morning coffee, terminal. Your local notes and Joplin are now out of sync in both directions: new notes were created on your phone, and some notes were modified on either side.

First, you pull the notes created on your phone:

```bash
$ kl joplin import
```

Then you push any local notes that are still missing from Joplin:

```bash
$ kl joplin export
```

Finally, you reconcile the notes that were *modified* on either side:

```bash
$ kl joplin merge
```

Note that `merge` keeps the newest version of each note, with no conflict resolution: if the same note was modified both locally and in Joplin, the older version is overwritten.

That's precisely why you also manage `$K_DIR` with git — Joplin synchronization is nice, but a proper history is safer:

```bash
$ git add . && git commit -m "Update and add new JavaScript notes"
```

Last, some housekeeping:

```bash
$ kl joplin clean
$ kl clean
```

## A few weeks later — Memorizing for good

As predicted, you've reopened the `array comparison` note three times already. Looking things up in your notes is fast, but some concepts you'd rather know by heart. Time to use spaced repetition.

Anki decks are defined by plain files in `$K_DIR` named `anki_export_<deck_name>`, each containing a list of note filenames (one per line). You create a deck with your two JavaScript notes:

```bash
$ echo "20240315100000.md" > anki_export_javascript
$ echo "20240315120000.md" >> anki_export_javascript
```

Then you export it:

```bash
$ kl anki export
```

This creates `anki_cards_javascript.apkg` in your export directory (`/tmp/knowledge-export/` by default, configurable with `dirExport`). Import it into Anki, and a few minutes of daily review will move those concepts from your notes into your head — no more reopening the same note every week.
