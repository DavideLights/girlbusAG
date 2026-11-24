# 🚍 GIRLBUS: NEON STOP (`girlbusAG` — v2.4 Build Stabile Pre-Espansione)
> **2D Pixel-Art Rogue-lite × Bullet-Hell × Tactical Dating Arcade**
> *Katana ZERO* (Personaggio a piedi, Chronos, Hit-Stop, VHS Rewind) × *PokéBonk* (Arena Continua 8–10 min, Orde Esponenziali) × *Stardew Valley* & *Pragmata* (Micro-Minigiochi In-Combat da 1.5–3s) × *GTA* (Polizia 1★–5★ Risk/Reward).

---

## 📂 Documentazione Ufficiale di Progetto

1. **[01_GAME_DESIGN_DOCUMENT.md](./01_GAME_DESIGN_DOCUMENT.md)** — **Game Design Document (GDD v2.4)**
   * **Protagonista & Controlli ("Zero Hand-Lift")**. Avatar a piedi agile stile *Katana ZERO* (`WASD` movimento + `Spazio`/`Click DX` Dash i-frames, `Shift`/`Q` Chronos Hold o Toggle, `Click SX` mira/lancio + azioni minigiochi e combo con opzione **Controllo frecce** su **Pad nel riquadro** o **Attorno alla fermata**, `Frecce/IJKL` fallback tastiera, Gamepad Twin-Stick).
   * **Core Loop & Profondità Tattica**: 4 archetipi di Fermate/Ragazze (*Emo*, *Studentessa*, *Kawaii*, *Cyber-Gyaru*), *Perfect Drop (`<= 24px`)*, vetrata scudo posteriore, *Bullet Graze* & *Dash-Parry* riduzione cooldown, *Stop Recall (`50%` rimborso + EMP)*, *Rizz Fever Streak (`x2`→`x5`)*, *Bus Slipstream* + catture *Al Volo*, *Bassline Magnet-Beat*, prioritizzazione click su ragazze/minigiochi e coda Level-Up anti-interruzione.
   * **Sistema Polizia ("Wanted 1★–5★")**: Bersagli **VIP Off-Limits** e **Zone Divieto di Sosta (ZTL)** pagano **3x Punti/Soldi** ma alzano le stelle Polizia (Vigili → Volanti → Barriere laser 3★ → SWAT anti-fermata 4★ → Elicottero con faro 5★), aumentando il moltiplicatore globale fino a **`3.5x XP / 4x Cash`**.
   * **I 6 Boss Gatekeeper, Overdrive & Boss Rush**: *3-Stop Break* su 4 Boss per run (*La Migliore Amica*, *L'Amante*, *La Professoressa*, *La Dirigente*, più *La Rettrice* / *La Mamma*), oltre a **Overdrive ciclico** post-vittoria e modalità **Boss Rush 6/6**.
   * **Costumi Pixel-Art & Challenge**: **10 Challenge** sbloccabili con premi visivi in `SFIDE & SKIN`, **5 Skin di Alfredo** (*Base*, *Neon Drip*, *Alfredo Banana*, *Aletta di Pollo*, *Cyber-Samurai Zero*), **Autobus Dorato Drift** e **Titoli Profilo** equipaggiabili.

2. **[02_TECH_STACK_AND_ARCHITECTURE.md](./02_TECH_STACK_AND_ARCHITECTURE.md)** — **Technical Stack, Risk Analysis & Architecture Blueprint (v2.4)**
   * Stack ufficiale **`Go + Ebitengine v2.8`**, pool proiettili *Zero-GC* pre-allocato (`4096` slot), shader **Kage** unico (`Pixel Crisp` / `Warm CRT` / `VHS Noir` + Chronos/Rewind), audio procedurale Synthwave + 27 SFX senza file esterni e astrazione `MapDef` pronta per nuove arene.

3. **[03_ART_STYLE_AND_SPRITE_BIBLE.md](./03_ART_STYLE_AND_SPRITE_BIBLE.md)** — **Art Direction & Sprite Bible (v1.1)**
   * Direzione *Vibrant Pop-Urban Chibi Pixel Art*, palette unificata, regole `1px` senza dithering sporco, costumi 32×32 e bus 96×48 generati via pipeline Python (`tools/`).

---

## ✅ Stato Attuale vs Prossima Fase (Espansione P2)

* **Completato & Verificato (Pre-Espansione)**:
  * Intero gameplay loop singolo giocatore, 6 Boss, **Chiavi Boss & Forzieri Blindati Neon** (sui `ChestSpots` di `MapDef`), Polizia 1★–5★, Distributori Neon, Slot 777, 6 Reliquie, Tutorial interattivo a 4 step, Overdrive ciclico, Boss Rush, P0 Ergonomia, P1 Meccaniche Loop e test automatici superati.
  * Dalla home, **Centro Neon** è giocabile. **Spiaggia** è un'anteprima bloccata e **Upcoming soon** annuncia le mappe future. Usa i pulsanti del selettore, la rotellina sopra l'anteprima, `A/D`, le frecce o la croce direzionale. **Impara a giocare** apre la guida e il tutorial interattivo.
  * La schermata titolo, i menu e i pannelli di partita condividono la grafica della segnaletica urbana. L'HUD mostra salute attuale e massima, Chronos, ricercato, XP e cooldown delle fermate. I riquadri dei minigiochi restano ancorati al lancio.
  * Il quartiere ha 13 sprite rifiniti, luci più localizzate e riflessi sul canale. Corsa, dash, posa di lancio, Perfect Drop, scudi dei boss e passaggio del bus hanno animazioni ed effetti dedicati.
* **Prossima Fase (Espansione Contenuti P2)**:
  * **Nuove Mappe Sbloccabili & Nuovi Archetipi**: *Ragazza Musicista* (Virtuoso Hero 4 colonne + Time-Freeze), *Ragazza Palestrata* (Max Trazioni / Pull-Up Mash `LMB`), *Ragazza Pallavolista* (Mini-Rally Volley) + concept Dual-Click (`LMB + RMB`).
  * **4 Eventi Dinamici di Quartiere**, **Co-op 2–4 Giocatori ("Caos Condiviso")** e **HUD Touch + Build Android**.

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
