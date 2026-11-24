# 🚍 GIRLBUS: NEON STOP (`girlbusAG` — v2.4 Build Stabile Pre-Espansione)
> **2D Pixel-Art Rogue-lite × Bullet-Hell × Tactical Dating Arcade**
> *Katana ZERO* (Personaggio a piedi, Chronos, Hit-Stop, VHS Rewind) × *PokéBonk* (Arena Continua 8–10 min, Orde Esponenziali) × *Stardew Valley* & *Pragmata* (Micro-Minigiochi In-Combat da 1.5–3s) × *GTA* (Polizia 1★–5★ Risk/Reward).

---

## 📂 Documentazione Ufficiale di Progetto

1. **[01_GAME_DESIGN_DOCUMENT.md](./01_GAME_DESIGN_DOCUMENT.md)** — **Game Design Document (GDD v2.5)**
   * **Protagonista & Controlli ("Zero Hand-Lift")**. Avatar a piedi agile stile *Katana ZERO* (`WASD` movimento + `Spazio`/`Click DX` Dash i-frames, `Shift`/`Q` Chronos Hold o Toggle, `Click SX` mira/lancio + azioni minigiochi e combo con opzione **Controllo frecce** su **Pad nel riquadro** o **Attorno alla fermata**, `Frecce/IJKL` fallback tastiera, Gamepad Twin-Stick).
   * **Core Loop & Profondità Tattica**: 4 archetipi di Fermate/Ragazze (*Emo*, *Studentessa*, *Kawaii*, *Cyber-Gyaru*), *Perfect Drop (`<= 24px`)*, vetrata scudo posteriore, *Bullet Graze* & *Dash-Parry* riduzione cooldown, *Stop Recall (`50%` rimborso + EMP)*, *Rizz Fever Streak (`x2`→`x5`)*, *Bus Slipstream* + catture *Al Volo*, *Bassline Magnet-Beat*, prioritizzazione click su ragazze/minigiochi e coda Level-Up anti-interruzione.
   * **Sistema Polizia ("Wanted 1★–5★")**: Bersagli **VIP Off-Limits** e **Zone Divieto di Sosta (ZTL)** pagano **3x Punti/Soldi** ma alzano le stelle Polizia (Vigili → Volanti → Barriere laser 3★ → SWAT anti-fermata 4★ → Elicottero con faro 5★), aumentando il moltiplicatore globale fino a **`3.5x XP / 4x Cash`**.
   * **I 6 Boss Gatekeeper, Overdrive & Boss Rush**: *3-Stop Break* su 4 Boss per run (*La Migliore Amica*, *L'Amante*, *La Professoressa*, *La Dirigente*, più *La Rettrice* / *La Mamma*), oltre a **Overdrive ciclico** post-vittoria e modalità **Boss Rush 6/6**.
   * **Costumi Pixel-Art & Challenge**: **10 Challenge** sbloccabili con premi visivi in `SFIDE & SKIN`, **5 Skin di Alfredo** (*Base*, *Neon Drip*, *Alfredo Banana*, *Aletta di Pollo*, *Cyber-Samurai Zero*), **Autobus Dorato Drift** e **Titoli Profilo** equipaggiabili.

2. **[02_TECH_STACK_AND_ARCHITECTURE.md](./02_TECH_STACK_AND_ARCHITECTURE.md)** — **Technical Stack, Risk Analysis & Architecture Blueprint (v2.5)**
   * Stack ufficiale **`Go + Ebitengine v2.8`**, pool proiettili *Zero-GC* pre-allocato (`4096` slot), shader **Kage** unico (`Pixel Crisp` / `Warm CRT` / `VHS Noir` + Chronos/Rewind), audio procedurale Synthwave + 27 SFX senza file esterni e astrazione `MapDef` pronta per nuove arene (outdoor e indoor).

3. **[03_ART_STYLE_AND_SPRITE_BIBLE.md](./03_ART_STYLE_AND_SPRITE_BIBLE.md)** — **Art Direction & Sprite Bible (v1.2)**
   * Direzione *Vibrant Pop-Urban Chibi Pixel Art*, palette unificata, regole `1px` senza dithering sporco, costumi 32×32, bus 96×48 generati via pipeline Python (`tools/`) e specifiche per Portali Neon/VHS e mappe indoor.

---

## ✅ Stato Attuale vs Idee & Espansioni Future (P2/P3)

* **Completato & Verificato (Build Attuale v2.4+)**:
  * Intero gameplay loop singolo giocatore, 6 Boss (con **Boss Arena Spotlight**, diradamento folla, limite Polizia e banner gimmick tattico), **12 Card Level-Up** con filtraggio intelligente anti-ripetizione, **Chiavi Boss & Forzieri Blindati Neon**, Polizia 1★–5★, Distributori Neon anti-abuso (cooldown e prezzi crescenti, reliquie bloccate a 6/6), **Multi-fermata a cariche fluide** (`0.22s` burst delay + Grace Drain `0.25x` + Express Chain `+35%`), **Area Bus / ZTL ripotenziate**, Slot 777, Tutorial interattivo a 4 step, Overdrive ciclico, Boss Rush e suite di test automatici superata.
  * **4 Eventi Dinamici di Quartiere (Implementati)**: *Ingorgo ZTL 4x* (`22s`, aree Bus dorate 4x), *Happy Hour Neon* (`20s`, `-35%` Shop + 1 Spin Slot Gratis), *Flashmob Campus* (`25s`, ondata concentrata di 1 archetipo) e *Blackout Neon* (`18s`, `+50%` Chronos immediato, luci localizzate ad alto contrasto e `+50%` Score/XP a fine evento).
  * **Onboarding Primi 60s, Personalità Arcade & Riepilogo Build/Seed (Implementati)**: rimborso `-40% CD` su lancio a vuoto iniziale e bonus primo imbarco (`+15 XP` & fermata subito pronta) entro i primi `60s`, micro-battute arcade non bloccanti per i 6 Boss / 4 Synergy / fine corsa e schermata **Riepilogo Build & Seed (`#XXXXXX`)** con breakdown per archetipo, Reliquie e Upgrade attivi.
  * Dalla home, **Centro Neon** è giocabile. **Spiaggia** è un'anteprima bloccata e **Upcoming soon** annuncia le mappe future. Usa i pulsanti del selettore, la rotellina sopra l'anteprima, `A/D`, le frecce o la croce direzionale. **Impara a giocare** apre la guida e il tutorial interattivo.
  * La schermata titolo, i menu e i pannelli di partita condividono la grafica della segnaletica urbana. L'HUD mostra salute attuale e massima, Chronos, ricercato, XP e cooldown delle fermate. I riquadri dei minigiochi restano ancorati al lancio.
  * Il quartiere ha 13 sprite rifiniti, luci più localizzate e riflessi sul canale. Corsa, dash, posa di lancio, Perfect Drop, scudi dei boss e passaggio del bus hanno animazioni ed effetti dedicati.
* **Idee & Espansioni Future da Approfondire (Non Ancora Implementate — Nessuna Scartata)**:
  * **P2 — Profondità Roguelite (Build Diversity, Agenzia & Difficoltà)**: Card/Reliquie trasformative con trade-off forti, Sinergie *Duo* tra archetipi, Evoluzioni al Rank Max, **Reroll a pagamento in Cash** (e Banish/Lock) sui Level-Up, Stili di partenza opzionali legati ai Costumi (con toggle *Solo Cosmetico*), Challenge con sblocchi di gameplay nel pool in-run e **Grado Traffico / Rush Level 1–5** post-vittoria.
  * **P3 — Nuove Mappe Sbloccabili, Archetipi, Co-op & Mobile**: *Spiaggia*, **Atrio Universitario / Campus (Indoor con Lucernario & Bus da Portali Neon/VHS)**, *Palestra* e *Club* + *Ragazza Musicista* (Virtuoso Hero 4 colonne + Time-Freeze), *Ragazza Palestrata* (Max Trazioni / Pull-Up Mash `LMB`), *Ragazza Pallavolista* (Mini-Rally Volley), concept Dual-Click (`LMB + RMB`), **Co-op 2–4 Giocatori ("Caos Condiviso")** e **HUD Touch + Build Android**.

---

## 🚀 Come Eseguire, Testare e Compilare

```powershell
# Avvio Desktop (richiede Go 1.22+)
go run .

# Build Windows (dist/girlbus-premium.exe)
go build -buildvcs=false -ldflags="-H windowsgui -s -w" -o dist\girlbus-premium.exe .

# Analisi statica e suite di test
go vet ./...
go test -count=1 -v ./...

# Build WebAssembly per browser (dist/web/girlbus.wasm)
$env:GOOS = "js"
$env:GOARCH = "wasm"
go build -ldflags="-s -w" -o dist\web\girlbus.wasm .
```
