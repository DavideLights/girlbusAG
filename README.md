# 🚍 GIRLBUS: NEON STOP (`girlbusAG` — Specifica v2.0 Allineata)
> **2D Pixel-Art Rogue-lite × Bullet-Hell × Tactical Dating Arcade**
> *Katana ZERO* (Personaggio a piedi, Chronos, Hit-Stop, VHS Rewind) × *PokéBonk* (Arena Continua 8–10 min, Orde Esponenziali) × *Stardew Valley* & *Pragmata* (Micro-Minigiochi In-Combat da 1.5–3s) × *GTA* (Polizia 1★–5★ Risk/Reward).

---

## 📂 Documentazione Ufficiale di Progetto

1. **[01_GAME_DESIGN_DOCUMENT.md](./01_GAME_DESIGN_DOCUMENT.md)** — **Game Design Document (GDD v2.0)**
   * **Protagonista & Controlli**: Avatar a piedi agile stile *Katana ZERO* (movimento + dash i-frames), lancio mirato a distanza delle Fermate (ogni Fermata ha **Gittata di Lancio** e **Raggio d'Azione** specifici), input unificati a 4 azioni per Desktop/Web/Android Touch, e Autobus Neon che sfreccia a raccogliere le ragazze a combo chiusa.
   * **Sistema Polizia ("Wanted 1★–5★")**: Bersagli **VIP Off-Limits** e **Zone Divieto di Sosta (ZTL)** pagano **3x Punti/Soldi** ma fanno arrivare la Polizia (Vigili → Volanti → Posti di Blocco → SWAT → Elicottero). **Più Stelle Polizia hai, più sale la difficoltà del bullet-hell e più aumenta il Moltiplicatore Globale su XP e Soldi (`fino a 3.5x XP / +300% Soldi`)!**
   * **Struttura Partita (8–10 min) & I 6 Boss Gatekeeper**: Arena continua evolvibile in cui il passaggio da un Livello di Difficoltà al successivo avviene battendo **3–4 Boss estratti a sorte dal Pool dei 6** tramite la meccanica **Boss 3-Stop Break**:
     1. *La Migliore Amica* (Gelosia / Spinge le tipe fuori dal raggio fermata)
     2. *L'Amante* (Doppio gioco / Cloni olografici esca)
     3. *La Professoressa* (Laser geometrici & Esame punitivo su errore combo)
     4. *La Dirigente Aziendale* (Scorta corporate & Muri di contratti)
     5. *La Rettrice* (Boss Finale A: Legge Marziale ZTL + Polizia istantanea)
     6. *La Mamma* (Boss Finale B: Ciabatta Boomerang a ricerca & "In Punizione!" che disattiva un perk)
   * **Puro Rogue-like ("Castrazione" Iniziale Fissa) & Sbloccabili via Challenge**: Si riparte **SEMPRE 100% castrati** ad ogni partita. **XP** = Level-up in-run; **Soldi** = Spesi solo nei Distributori Neon in-run (upgrade rari, cure, tangente Polizia). Tra le partite si sbloccano **Nuove Mappe, Skin (Protagonista/Pensiline/Bus) e Fermate** completando **Challenge**!
   * **Co-op 2–4 Giocatori ("Caos Condiviso")**: Stelle Polizia globali, Soldi individuali, **Rianimazione lanciando una Fermata sul compagno KO**, Bolla Chronos Locale e Linea Express (*Ghost Bus Rush*).

2. **[02_TECH_STACK_AND_ARCHITECTURE.md](./02_TECH_STACK_AND_ARCHITECTURE.md)** — **Technical Stack, Risk Analysis & Architecture Blueprint (v2.0)**
   * **Analisi del Rischio Rust vs Go**: Perché in un rogue-lite ad alta interazione (*Fermate ↔ Ragazze ↔ Polizia ↔ Boss*) il Borrow Checker di Rust costringe a usare indici manuali e code di eventi che aumentano la complessità del codice e il rischio di bug logici, incoronando **`Go + Ebitengine`** come **Scelta Ufficiale #1**.
   * **Zero-GC Bullet-Hell in Go**: Come gestire 5.000+ proiettili in `~120 KB` di RAM contigua e `0.00 ms` di pausa Garbage Collector usando slice pre-allocati di struct (`[]Bullet`).
   * **Unica Codebase (Linux + Web Wasm + Android)** + **Shader Kage** (scritti direttamente in sintassi Go!) + Albero delle directory piatto e leggibile.
