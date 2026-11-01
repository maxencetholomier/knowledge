# System Architecture

Knowledge operates as a local-first system. The cloud integration is the "icing on the cake".

On your computer, it manages markdown notes in your filesystem (creation, research, deletion, etc.) and synchronizes them bidirectionally (import, export and merge) with two applications, each serving a different purpose:

- [Joplin](https://github.com/laurent22/joplin), to read and write the notes on every device
- [Anki](https://apps.ankiweb.net/), to learn selected notes through spaced repetition

Both work the same way: the local notes stay the reference, and edits made on the phone come back through the application's own sync and a merge on the computer.

```mermaid
%%{init: {'flowchart': {'rankSpacing': 80, 'nodeSpacing': 60}}}%%
flowchart TB
    subgraph Cloud
        JC(Joplin Cloud)
        GH(Github/Gitlab)
        AW(AnkiWeb)
    end
    subgraph Phone
        JM(Joplin Mobile)
        AM(AnkiDroid)
    end
    subgraph Computer
        JD(Joplin Desktop)
        G(Git)
        AD(Anki Desktop)
        K(Knowledge CLI)
    end
    JC <--> JM
    AW <--> AM
    JC <--> JD
    GH <--> G
    AW <--> AD
    G --- K
    JD <--> K
    AD <--> K
    classDef joplin fill:#dbeafe,stroke:#2563eb,color:#1e3a8a
    classDef anki fill:#ffedd5,stroke:#ea580c,color:#7c2d12
    classDef git fill:#dcfce7,stroke:#16a34a,color:#14532d
    class JM,JC,JD joplin
    class AM,AW,AD anki
    class GH,G git
```
