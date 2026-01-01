# Problems with Classic Note-Taking

The "classic note-taking" method consists of storing information in a set of files distributed across multiple folders and subfolders.
For organizing small amounts of information, this method is perfectly suitable.

However, it becomes difficult to apply when storing large quantities of data. Let's look at an example case together.

## Example Case

In addition to your work, you regularly research new consumer technologies.
You learned that your WIFI-6 router uses a new signal processing protocol called MU-MIMO.
You would like to store this information in your notes, since this information was complicated to find and you don't want to waste time searching again.

You open your directory and search for all existing notes that might contain this information.

```
shell> cd ~/my_notes
shell> tree
.
├── physique
│   └── traitement_du_signal
│       └── multiple_outputs.md
└── telecom
    └── protocole_de_communication
        └── wifi.md
```

Two notes could be suitable: **multiple_outputs.md** and **wifi.md**.

Since most files are over 1000 lines long, it takes you several minutes to identify where to place the information in each of these notes.

The formats of the two notes being very different, you seize the opportunity to normalize these two notes. You apply the same title and subtitle structure and add a table of contents to **wifi.md** which didn't have one.

Now you must make a choice. In which file will you classify the information? The most relevant file?

Being someone very busy, you doubt that in 10 weeks you will still be able to remember where you stored the information.

To avoid choosing, you consider putting a link from wifi.md to multiple_outputs.md. However, the links will be broken during folder reorganization. Plus you're short on time.

What to do?

⏰ Ugh, your phone rings - it's already time for pottery class. ⏰

You postpone adding this note to later.

## Identifying the Problems

To summarize, here are the main problems that can be encountered:

- Some notes can belong in multiple folders. However, duplicating or making a choice is not a good idea.
- Some notes are too long. It becomes more difficult to identify useful information.
- The absence of logical links between notes doesn't allow finding related information.
- If links exist, folder reorganization makes them complicated to manage.
- From experience, people tend to want to standardize note formats. This is a waste of time.


The [Zettelkasten Method](zettlekasten.md) helps solve these problems in part.
