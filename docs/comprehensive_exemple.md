# Comprehensive Example

You've just gotten a new job as a React developer.

Despite your 10 years of experience, it's been a while since you've worked on JavaScript,
and as long as you remember you always forgot syntax subtleties, like what is `falsy` or `array comparison`.


Going to the internet for each recurring question wastes time.
And time is money...

## Arriving at the office

You take your coffee and start launching a terminal.

Out of curiosity, you want to see how many notes you already have:

```bash
$ kl list | wc -l
3477
```

That's a lot of knowledge. Ten years of experience accumulates significant information.
Hopefully you've organized it well and haven't lost any knowledge.

This makes you happy before starting the journey!

## Start working

Now you want to know if you already have notes about the things you always forgot.

```bash
$> kl find javascript
1 JavaScript Array Methods
2 JavaScript falsy values
3 ES6 Features Overview
....
```

Note on `falsy` exists but nothing on array comparison...

So you open the note on `falsy` values.

```bash
kl edit 2
```

You realize that you missed the `""` value. So you update the note and quit your editor.

Then you go on the internet and review `array comparison`, again...

You create the note using the following command

```bash
kl new
```

This creates `20240315100000.md` and opens your editor.

You Add:

```markdown
# JavaScript Array Comparison

- Arrays are objects, so `==` and `===` compare references, not content
- Use helper functions or loops to compare array content
- JSON.stringify() for simple arrays (has limitations)
```

Your knowledge is now solid.

You can come back when you forget this notion again, probably next week...

## Leaving the office

On your commute home (using a bus) you do not have a lot of things to do.
Going on Instagram would be a waste of time.

It would be nice to review what notes have been added today.

So before leaving you export the new notes on Joplin Desktop.

```bash
kl joplin export
```

You are ready to go home!

## Go to work Again

During your commute back you try to learn what's been added in JavaScript during the 10 years
you haven't touched it.

You realize that there are a lot of things you've missed.
So you create and update a lot of your notes.

Arriving at work, you take your morning coffee and start importing the new notes:

```bash
kl joplin import
```

You also export the notes that you've done on your computer to your phone

```bash
kl joplin export
```

You also want to have the notes you've modified:

```bash
kl joplin merge
```

That's nice. The only thing missing is to make sure the notes are properly saved.

Joplin synchronization is nice but it's definitely not perfect.

You know it and that's why you are managing your $K_DIR with git

```bash
git add . && git commit -m "update and add new Javascript Note"
```
Then you clean your notes :

```bash
kl joplin clean
kl clean
```

## Preparing for Interview

A few weeks later, you realize you've accumulated quite a bit of JavaScript knowledge. With job interviews coming up, you want to use spaced repetition to memorize the key concepts.

First, you create an Anki deck definition for your JavaScript notes:

```bash
echo "20240315100000.md" > anki_export_javascript
echo "20240315120000.md" >> anki_export_javascript
```

Then you export your knowledge to Anki format:

```bash
kl anki export
```

This creates `anki_cards_javascript.apkg` in your export directory, which you can import into Anki for spaced repetition study sessions.

Now you can review your JavaScript knowledge daily using Anki's proven spaced repetition algorithm, ensuring you remember the concepts during your interviews.
