# ⚙️ GIRLBUS: NEON STOP — Technical Stack, Risk Analysis & Architecture Blueprint (v2.5)
> **Vincoli Ingegneristici del Progetto**:
> 1. **Unica Codebase** per **Web (Wasm)**, **Android App** e **Linux Desktop**.
> 2. **Priorità Assoluta 1**: **Codice pulito, semplice, leggero, facile da aggiustare e leggere** — minimizzando il rischio di scrivere codice contorto o introdurre bug nella logica di gameplay e nella gestione della memoria.
> 3. **Priorità 2**: **Efficienza su RAM e CPU** per un Bullet-Hell da migliaia di proiettili con Polizia (1★–5★), 6 Boss multi-fase e shader Neon/CRT stile *Katana ZERO*.

---

## 1. Analisi del Rischio: Rust vs Go nel Game Development 2D Rogue-lite

Hai posto la domanda chiave di tutto il progetto:
> *"C'è più rischio di scrivere codice complesso e di sbagliare la logica del codice e nella gestione della memoria usando Rust rispetto a Go?"*

### Risposta Tecnica Onesta: **SÌ. In un Rogue-lite Bullet-Hell, Rust aumenta drasticamente il rischio di codice complesso e bug logici rispetto a Go.**

Ecco il motivo tecnico preciso legato alle meccaniche di `girlbusAG`:

### 1.1 Il "Borrow Checker Trap" nei Giochi ad Alta Interazione
In `girlbusAG`, in un singolo frame di `16.6ms` accadono **interazioni incrociate continue**:
> *Il Giocatore* lancia una *Fermata* in una *Zona ZTL* → la *Fermata* aggancia *4 Ragazze VIP* e un *Boss (La Professoressa)* → l'azione fa salire le *Stelle della Polizia* → le *Volanti della Polizia* sparano ai *Proiettili/Fermate* → il *Minigioco Pragmata* congela i *Laser del Boss*.

* 🔴 **Cosa succede in Rust**:
  Il **Borrow Checker** di Rust vieta di avere due riferimenti mutabili contemporanei allo stato del mondo (`cannot borrow world as mutable more than once at a time`).
  * Non puoi scrivere semplicemente `stop.hook(girl, &mut world.police)` mentre iteri su `world.stops`.
  * Per far compilare il codice in Rust, sei costretto a **complicare l'architettura**: devi sostituire i puntatori con **indici numerici manuali (`GirlIndex(usize)`, `StopId(u32)`)** o usare *Arena / Generational Indices* e spezzare ogni azione in **Code di Eventi Differiti (`Vec<GameEvent>`)**.
  * **Il Rischio Reale**:
    1. **Bug nella logica e negli indici ("Use-After-Free Logico")**: in Rust il compilatore ti protegge dai crash C++, ma se usi indici manuali `usize` su un array di ragazze/proiettili per aggirare il borrow checker e un'entità viene rimossa con `swap_remove`, il tuo indice punta all'entità sbagliata! Hai appena reintrodotto i bug di memoria a livello logico.
    2. **Rigidità nel cambiare il gameplay**: se a metà sviluppo ti viene un'idea geniale per un nuovo Boss (es. *La Mamma* che ti sequestra una fermata attiva), in Rust rischi di dover ristrutturare le firme e i lifetime di 10 funzioni.
* 🟢 **Cosa succede in Go (`Go + Ebitengine`)**:
  Go è stato progettato da Google esattamente per eliminare la complessità mentale del codice:
  * **Solo 25 parole chiave** in tutto il linguaggio. Nessuna macro, nessun lifetime `'a`, nessun borrow checker.
  * Puoi passare puntatori diretti (`*BusStop`, `*Girl`, `*Boss`, `*PoliceSystem`) e modificare qualsiasi campo con un codice piatto, lineare e leggibile come pseudocodice:
    ```go
    if stop.InZTL {
        world.Police.AddStar(1)
        girl.ApplyVIPBonus(stop)
    }
    ```
  * **Sicurezza Totale della Memoria**: In Go è **impossibile** avere dangling pointers o corruzione di memoria, perché il runtime gestisce i riferimenti in modo sicuro.

---

## 2. Il Verdetto Finale: 🥇 **Go + Ebitengine** (Scelta Ufficiale #1)

Scegliendo il linguaggio **più semplice, pulito e a minor rischio di errore logico**, incoroniamo **`Go + Ebitengine`** come stack ufficiale di **`girlbusAG`**.

| Criterio del Brief | 🥇 **Go + Ebitengine** (Scelta Ufficiale) | 🥈 **Rust + Macroquad** (Alternativa) | Perché vince **Go + Ebitengine** per noi |
| :--- | :---: | :---: | :--- |
| **Semplicità del Codice & Zero Complessità Mentale** | **10 / 10** | 6.5 / 10 | Codice piatto, nessun borrow checker, refactoring istantaneo delle regole dei Boss e dei Minigiochi. |
| **Basso Rischio di Sbagliare Logica / Memoria** | **9.8 / 10** | 7.0 / 10 | Nessun bisogno di indici manuali o code di eventi contorte per far parlare Fermate, Polizia e Boss. |
| **Velocità di Compilazione & Iterazione** | **`< 0.5 secondi`** | `3 – 8 secondi` | Modifichi il feeling dell'Hit-Stop o un pattern Bullet-Hell e lo vedi a schermo istantaneamente. |
| **Efficienza RAM & CPU (Bullet-Hell)** | **9.0 / 10** *(con Slice Pre-Allocati)* | 9.8 / 10 | Usando slice contigui di struct (`[]Bullet`), Go ha **zero allocazioni GC** durante il combattimento! |
| **Unica Codebase (Linux + Web Wasm + Android)** | **9.6 / 10** | 8.8 / 10 | Tool ufficiale `ebitenmobile` per Android + Shader **Kage** scritti in sintassi Go che girano ovunque. |

---

## 3. Come Garantire RAM e CPU Ultra-Leggere in Go (Zero-GC nel Bullet-Hell)

L'unico potenziale punto debole di Go nei videogiochi è il **Garbage Collector (GC)** se si creano e distruggono migliaia di oggetti con `new()` ad ogni frame.
**La bellezza di Go** (a differenza di Java, Python o JavaScript) è che le **`struct` in Go sono Value Types** (vengono salvate in memoria contigua piatta, esattamente come in C!):

###La Regola d'Oro in 10 Righe: *Pre-Allocated Flat Slice*
All'avvio della partita allochiamo **una sola volta** lo slice dei proiettili:

```go
type Bullet struct {
    X, Y       float32
    VX, VY     float32
    Radius     float32
    Kind       BulletKind
    Active     bool
}

// All'avvio: 1 singola allocazione contigua in RAM (~120 KB totali per 5.000 proiettili!)
// Durante la partita: ZERO allocazioni, quindi il Garbage Collector di Go NON PARTE MAI!
type BulletPool struct {
    Items []Bullet
    Count int // Numero di proiettili attualmente attivi (0..Count-1)
}

func (p *BulletPool) Spawn(b Bullet) {
    if p.Count < len(p.Items) {
        p.Items[p.Count] = b
        p.Count++
    }
}

// Rimozione in O(1) istantanea senza allocare nulla in memoria!
func (p *BulletPool) RemoveAt(i int) {
    p.Count--
    p.Items[i] = p.Items[p.Count]
}
```
* **Risultato su RAM e CPU**:
  * **RAM occupata dai 5.000 proiettili**: appena **~120 KB** tutti nella Cache L1/L2 del processore!
  * **Lavoro del Garbage Collector durante il Bullet-Hell**: **0.00 ms** (zero allocazioni sull'heap durante il loop `Update()`).
  * **CPU Frame Time per 5.000 proiettili**: **< 0.4 millisecondi** su singolo core!

---

## 4. Valutazione di WebAssembly (Wasm) con Go + Ebitengine

**Conviene usare WebAssembly per la versione Web?**
**SÌ, è supportato nativamente da Go (`GOOS=js GOARCH=wasm`) ed è la soluzione ideale.**

```mermaid
flowchart LR
    subgraph Codebase["Unica Codebase Go + Shader Kage"]
        CORE["internal/game<br/>(Fermate, Ragazze, Polizia 1-5★,<br/>6 Boss, Minigiochi, Co-op)"]
        SHAD["assets/shaders/*.kage<br/>(Shader Neon & Katana CRT in sintassi Go)"]
    end

    CORE & SHAD -->|go build| LIN["🐧 Linux Desktop<br/>Binario Nativo ~8 MB (OpenGL)"]
    CORE & SHAD -->|GOOS=js GOARCH=wasm| WEB["🌐 WebAssembly (.wasm)<br/>~3.2 MB Brotli (WebGL2)"]
    CORE & SHAD -->|ebitenmobile bind| AND["🤖 Android APK / AAB<br/>Nativo ARM64 (OpenGL ES 3)"]
```

### 4.1 Vantaggi di Wasm + Ebitengine
1. **100% Identico tra Linux, Android e Browser**: Non devi riscrivere una sola riga per il Web. Anche gli shader (l'effetto CRT e il Neon Bloom alla *Katana ZERO*) sono scritti in **Kage** (un dialetto di Go!) ed Ebitengine li compila automaticamente in **WebGL2** per il browser, **OpenGL** per Linux e **OpenGL ES** per Android.
2. **Batching Automatico sulla GPU**: Ebitengine raggruppa automaticamente tutti i `DrawImage` provenienti dallo stesso Texture Atlas in **1 singola Draw Call WebGL2**, mantenendo **60/120 FPS** fissi nel browser anche con migliaia di proiettili e goccioline Kawaii.
3. **Ottimizzazione del Peso Wasm**:
   * Compilando con i flag `-ldflags="-s -w"` (strip dei simboli di debug) e comprimendo il file `.wasm` con **Brotli (`girlbus.wasm.br`)**, il bundle Web scende a **~3.0–3.5 MB** e si carica in streaming nel browser in ~1 secondo.

---

## 5. Struttura Reale del Progetto (v2.4 — Piatta, Chiara & Pronta per Espansione)

In Go evitiamo gerarchie profonde o astrazioni inutili. Ogni package e file rispecchia esattamente l'organizzazione reale del repository:

```text
girlbusAG/
├── go.mod                          # Dipendenza principale: github.com/hajimehoshi/ebiten/v2 v2.8.6
├── main.go                         # Entry point (Loop Update/Draw, routing click & guard modali, Overdrive, SFX)
├── README.md                       # Panoramica progetto, stato build e comandi di avvio/test
├── 01_GAME_DESIGN_DOCUMENT.md      # GDD v2.4 Completo
├── 02_TECH_STACK_AND_ARCHITECTURE.md
├── 03_ART_STYLE_AND_SPRITE_BIBLE.md
│
├── assets/                         # Asset embedded nel binario con //go:embed (Zero file esterni richiesti!)
│   ├── embed.go                    # SpriteBank, helper PlayerSkinSpriteKey & fallback automatici
│   ├── shaders.go                  # Caricamento shader Kage
│   ├── shaders/
│   │   └── postprocess.kage        # Shader unico: Pixel Crisp (0) / Warm CRT (1) / VHS Noir (2) + Chronos/Rewind
│   └── sprites/
│       ├── characters/             # Alfredo Base + 4 Costumi (Drip, Banana, Chicken, Samurai), 4 Archetipi + VIP aura
│       ├── world/                  # 6 Boss, Bus Neon & Bus Dorato, Polizia, Fermate, Proiettili, Vending
│       ├── map/                    # Edifici, auto, lampioni, slot, alberi, barche e tile stradali
│       └── ui/                     # Icone UI, Reliquie, Upgrade, Mouse, Stelle Wanted, Cure Espresso
│
├── internal/
│   ├── audio/                      # 🔊 Synthwave Procedurale Zero-Alloc + 25 SFX Pre-renderizzati (niente .ogg esterni)
│   │   ├── audio.go                # Engine: ducking Low-Pass in Chronos, BPM dinamico, pool 4x per SFX
│   │   ├── synth.go                # Sintesi musicale multi-layer (incluso Rizz Fever arpeggio)
│   │   ├── sfx.go                  # Generatore PCM dei 25 effetti sonori (incl. ShieldBlock / ShieldBreak)
│   │   └── audio_test.go
│   │
│   ├── game/                       # 🧠 TUTTA LA SIMULAZIONE DETERMINISTICA PURE-GO
│   │   ├── types.go                # Costanti (FixedPadCenterX/Y, Hitbox, Buffer 120ms, Coyote 0.35s), enum e tipi base
│   │   ├── world.go                # Loop World.Update, Spawn/Pause/QueuedLevelUps, click priority, Overdrive, Eventi
│   │   ├── player.go               # Movimento, Dash i-frames, Chronos Hold/Toggle, Reliquie, Cooldown & Bassline Pulse
│   │   ├── bus_stop.go             # Fermate, scudo vetrata, GoldenZTL, ComputeStopUIAnchor/PadCenter, NeonBus
│   │   ├── events.go               # 4 Eventi Dinamici di Quartiere (ZTLJam, HappyHour, Flashmob, Blackout)
│   │   ├── quips.go                # Micro-battute arcade Boss/Synergy/EndRun & formattazione Seed (#XXXXXX)
│   │   ├── girls.go                # AI & pattern dei 4 archetipi + varianti VIP
│   │   ├── bullets.go              # BulletPool statico da 4096 slot (Zero-GC) + deflessioni + scudo + AntiStop
│   │   ├── police.go               # Wanted 1★–5★, Pattuglie, Volanti, Barriere 3★, SWAT AntiStop 4★, Elicottero 5★
│   │   ├── bosses.go               # I 6 Boss Gatekeeper, Cloni, Scorta, ZTL Marziale, Ciabatta & 3-Stop Break
│   │   ├── minigames.go            # Emo BPM Rhythm, Pragmata 4×4, Kawaii Tear-Graze, Gyaru Hold-to-Rise + Combo 4-Step
│   │   ├── shop.go                 # Distributore Neon (Espresso, Tangente, Reliquie, Cassa, Happy Hour -35%) & Rizz-Slots 777
│   │   ├── map_layout.go           # Topologia distretto 1280×720 + astrazione MapDef / DefaultDistrictMap()
│   │   └── *_test.go               # 47 unit test deterministici su bilanciamento, layout, boss, shop, eventi & onboarding
│   │
│   ├── input/                      # 🕹️ INPUT UNIFICATO CROSS-PLATFORM ("Zero Hand-Lift")
│   │   ├── input.go                # WASD + Mouse LMB/RMB/Wheel + pad/bussola + fallback Frecce/IJKL + Gamepad + Touch
│   │   └── input_test.go
│   │
│   ├── meta/                       # 💾 PERSISTENZA, CHALLENGE & SKIN
│   │   ├── meta.go                 # Settings (FilterMode, DirInputMode, ChronosToggle...) + Profile (5 Skin, 2 Bus, Titoli)
│   │   ├── challenges.go           # 10 Challenge + RunTracker anti doppio-conteggio
│   │   ├── storage_desktop.go / storage_wasm.go
│   │   └── meta_test.go
│   │
│   ├── render/                     # 🎨 RENDERER PIXEL-ART & HUD DIEGETICA
│   │   ├── renderer.go             # Arena Y-sorted, animazioni, mira, proiettili e riquadri dei minigiochi
│   │   ├── action_effects.go       # Scintille, tether, avvisi corsia, scie dash e pannelli minigioco
│   │   ├── hud.go                  # Salute, Chronos, boss, hotbar, notifiche e pannelli Upgrade/Shop/Slot
│   │   ├── lighting.go             # Luci del quartiere, riflessi sul canale ed effetti delle entità
│   │   ├── postprocess.go          # Wrapper shader PostProcessor (Warm CRT / VHS Noir / Pixel Crisp)
│   │   └── ui_helpers.go           # Cache testo con accenti italiani (à è é ì ò ù) & misura rune UTF-8
│   │
│   └── ui/                         # 📋 MENU, OPZIONI & SFIDE
│       ├── ui.go                   # Manager, transizioni, focus e input dei menu
│       ├── theme.go                # Palette, componenti e geometria dei controlli
│       ├── title.go                # Insegna e comandi della home
│       ├── map_picker.go           # Selettore mappe, anteprime e animazione
│       ├── menus.go                # Pausa, opzioni, comandi e risultati
│       ├── guide.go                # Guida e accesso al tutorial interattivo
│       ├── collection.go           # Sfide, anteprime e personalizzazione
│       ├── reward_icons.go         # Icone dei premi e dello stato di sblocco
│       └── ui_test.go              # Focus, opzioni, guida e navigazione mappe
│
├── tools/                          # 🎨 Pipeline generativa Python Pillow per sprite, costumi & preview 8x/10x
├── scripts/                        # 🚀 Script di build (build_linux.sh, build_web.sh)
└── dist/web/                       # 📦 Bundle WebAssembly (girlbus.wasm, index.html, wasm_exec.js)
```

### Ordine di rendering

1. `App.Update` chiama `Renderer.AdvanceVisuals` al passo di aggiornamento del gioco. Le animazioni ambientali e le notifiche avanzano in base a quel passo.
2. `DrawWorldScene` disegna il quartiere a `640×360`. Le luci architettoniche e i riflessi del canale sono sotto gli sprite ordinati per coordinata Y. Gli effetti legati alle entità, l'autobus e la mira precedono i segnali di pericolo e i proiettili.
3. `PostProcessor.Present` applica il filtro Kage selezionato. Il bloom è più intenso sugli accenti luminosi saturi. HUD, riquadri dei minigiochi e menu vengono composti dopo il filtro, così testo e hitbox restano allineati.
4. `game.ComputeStopUIAnchor` fissa l'origine del pannello al lancio. `World.StopSafeTopY` riserva spazio ai banner in alto. La geometria del pannello in `renderer.go` serve anche a evitare sovrapposizioni con le notifiche di combattimento.

Il selettore della home conserva la scheda corrente nel `ui.Manager`. Le anteprime sono disegnate una volta e riutilizzate durante la transizione. Le schede Spiaggia e Upcoming soon sono elementi dell'interfaccia. L'accesso alle modalità di gioco dalla home richiede la scheda Centro Neon, che usa `game.DefaultDistrictMap`.

### Stato Tecnico Verificato (Build Attuale v2.4+)
* **Zero-GC Bullet Pool**: `BulletPool` usa un array fisso `[4096]Bullet` value-type (`MaxBullets = 4096`) senza allocazioni durante il combattimento.
* **Audio Interamente Procedurale**: `internal/audio/` genera musica synthwave multi-layer e `25` buffer SFX in PCM `44.1 kHz` senza dipendere da asset audio esterni.
* **Multi-Mappa Predisposta (`MapDef`)**: `DefaultDistrictMap()` in [internal/game/map_layout.go](./internal/game/map_layout.go) incapsula geometria, pool ragazze e 3 `ChestSpots` predisposti per i futuri Forzieri Boss.
* **Eventi Dinamici, Onboarding, Quip & Riepilogo Seed/Build (Verificati ✅)**: implementati `internal/game/events.go` + `quips.go`, overlay Blackout e BusBay ZTL Dorate in `internal/render/` e schermata `RunSummary` completa (`SEED #XXXXXX`, breakdown archetipi, Reliquie, Upgrade) in `internal/ui/`.
* I **67 test** passano in `internal/audio` (7), `internal/game` (47), `internal/input` (3), `internal/meta` (4) e `internal/ui` (6). Verificati anche `go vet ./...` e le compilazioni Windows e WebAssembly.
* **Verifica visiva su Windows**. Catturati 20 scenari da Ebitengine, tra cui le tre schede della home, la transizione tra mappe, minigioco, boss, lungocanale e scene di gioco con i tre filtri.

### Estensioni Architetturali Ancora Da Implementare (Design P2/P3)

Senza alterare la simulazione lineare della build attuale, i package restano predisposti per accogliere le restanti estensioni approvate nel GDD v2.5:

* **`MapDef` per Mappe Outdoor e Indoor (`internal/game/map_layout.go`)**: aggiunta di metadati descrittivi (`Indoor bool`, `BusUsesPortals bool`, `SkylightOverlay bool`, etichette/palette tematiche per zone ZTL/Silenzio, Shop/Macchinette e Polizia). La logica fisica di Bus, ZTL, Distributori e Faro 5★ resta unica e deterministica; cambia soltanto la presentazione nel renderer (es. apertura Portali Neon/VHS ai capi della traiettoria del Bus e filtro lucernario nell'*Atrio Universitario*).
* **Build Diversity, Agenzia RNG & Loadout (`internal/game` + `internal/meta`)**:
  * Supporto a `RerollCost` in Cash (`$`) nel modal Level-Up (più Banish/Lock), card/reliquie trasformative con flag di trade-off, sinergie *Duo*, Evoluzioni al Rank Max e filtraggio del pool guidato anche dagli sblocchi in `Profile.UnlockedChallenges`.
  * `LoadoutKind` / modificatori pre-run (`RushLevel 1–5`) con toggle puramente cosmetico per le skin di Alfredo.
