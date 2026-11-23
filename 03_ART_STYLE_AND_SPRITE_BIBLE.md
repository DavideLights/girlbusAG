# 🎨 GIRLBUS: NEON STOP — Art Direction & Sprite Bible (v1.0)
> **Sintesi Estetica**: **"Vibrant Pop-Urban Chibi Pixel Art"**
> Unione tra le meccaniche ad alta tensione di *Katana ZERO / PokéBonk* e uno stile visivo **solare, saturo, ironico e super-leggibile** ispirato alla pixel art urbana giapponese contemporanea (`peak-envref7`), alla densità satirica da metropoli pop (`peak-envref4`) e a sprite **Chibi 3/4 Top-Down ultra-espressivi** (`peak-npcref4`, `peak-npcref5`, `peak-npcref6`, `peak-npcref1`), con **Alfredo (versione base urban/casual)** come protagonista (`mainref-alfredo1/2`).

---

## 1. Contestualizzazione Completa della Cartella `ref/` (Gerarchia & Ruolo)

### 👑 A. Protagonista (`mainref-alfredo1.jpg`, `mainref-alfredo2.jpg`) — *Il Giocatore (Versione Base, NO Laurea)*
Le due foto servono **esclusivamente come riferimento fisionomico per il volto e il taglio di capelli di Alfredo** (ignorando completamente corona d'alloro, tesi e abito da laurea: lo sprite è un **personaggio base urbano quotidiano**).
* **Tratti Fisionomici & Outfit Base per lo Sprite di Alfredo (`32×32 px`)**:
  1. **Taglio di Capelli (`mainref-alfredo2.jpg`)**:
     * Capelli **castano scuro / neri**, taglio **corto e pulito** (sfumato/corto sui lati con orecchie ben visibili, frangia corta e ordinata sulla fronte).
     * L'assenza di cappelli o ciuffi lunghi rende la silhouette della testa di Alfredo super pulita e subito distinguibile dalle chiome voluminose/colorate delle ragazze (`peak-npcref4/5`).
  2. **Volto & Espressione**:
     * **Sopracciglia scure spesse e ben marcate** (tratto distintivo fondamentale per renderlo riconoscibile anche in `32×32 px`!).
     * Viso sbarbato e pulito, occhi scuri espressivi con highlight bianco, e **sorriso/ghigno sicuro di sé** (*"Rizz-Smirk"*).
  3. **Outfit Base Casual / Streetwear Urbano (stile `peak-npcref1` + `peak-npcref6`)**:
     * Look quotidiano fresco e leggibile in top-down: **giacca casual / felpa streetwear aperta** (es. blu navy/indaco o verde oliva/rosso urbano) sopra una **t-shirt chiara**, **jeans/pantaloni scuri** e **sneakers** con suola chiara (perfette per leggere il movimento dei piedi durante corsa e Dash).

---

### 🔥 B. Riferimenti Primari (`peak-*`) — *Massima Priorità Estetica*

#### 1. `peak-npcref4.jpg` & `peak-npcref5.jpg` — *Il Canone per le Ragazze (Archetipi & Boss)*
* **Cosa mostrano**:
  * [`peak-npcref4.jpg`](./ref/peak-npcref4.jpg): Griglia di 18 ragazze anime/idol in stile Chibi pixel art (codini, fiocchi, cuffie, abiti gothic-lolita, pastel-kawaii, maid, studentesse).
  * [`peak-npcref5.jpg`](./ref/peak-npcref5.jpg): Close-up ad alta leggibilità di una ragazza chibi dai capelli rosso-magenta con orecchie da volpe, salopette scura e maglia a righe.
* **Regole Auree estratte per `girlbusAG`**:
  * **Proporzioni Chibi / SD (`1.7–1.8 teste` di altezza)**: La testa occupa circa il **55–60% dell'altezza totale** dello sprite, il corpo e le gambe il restante **40–45%**.
  * **Perché funziona da dio in un Bullet-Hell**: La testa grande permette di riconoscere in `50 millisecondi` l'archetipo della ragazza (**Emo, Studentessa, Kawaii, Cyber-Gyaru** o uno dei **6 Boss**) grazie a tagli di capelli, fiocchi e colori saturi, mentre il corpo compatto mantiene la **hitbox fisica piccola e onesta** in mezzo ai proiettili!
  * **Occhi Anime Verticali (`2×3` o `2×4` px)**: Sclera bianca + pupilla colorata satura + `1px` di highlight bianco in alto.
  * **Outline Selettivo Scuro (`1px`)**: Contorno esterno deciso in viola-nero/marrone scurissimo (`#1E1224` / `#23161C`), mai grigio slavato, con linee interne morbide per ciocche e pieghe dei vestiti.
  * **Cluster Shading Pulito (Zero Dithering)**: Massimo **3 toni per colore** (Base, Ombra, Luce netta sui capelli). Nessun effetto sfumato sporco o rumore.

#### 2. `peak-npcref6.jpg` — *Prospettiva 3/4 Top-Down, Movimento & Props*
* **Cosa mostra**: [`peak-npcref6.jpg`](./ref/peak-npcref6.jpg) ritrae un personaggio chibi con cappello a punta, salopette verde e annaffiatoio tra vasi di piante.
* **Regole Auree estratte**:
  * **Angolazione Camera 3/4 Morbida**: Mostra perfettamente come tradurre le proporzioni chibi di `peak-npcref4/5` in **vista 3/4 da gioco top-down** (piedini arrotondati di `3×2` px, leggera rotazione del busto, braccia ben leggibili).
  * **Integrazione Personaggio-Prop**: Come il giardiniere tiene l'annaffiatoio, così i Boss terranno i loro oggetti iconici (es. *La Professoressa* con il goniometro laser, *La Mamma* con la ciabatta, *La Dirigente* con la valigetta).

#### 3. `peak-npcref1.png` — *Guardaroba Urbano Moderno & Archetipi Cittadini*
* **Cosa mostra**: [`peak-npcref1.png`](./ref/peak-npcref1.png) (*"MODERN PEOPLE - EVERYONE HERE HAS A STORY"*) presenta 7 cittadini contemporanei (donna in tailleur blu con borsa a tracolla, ragazzo con smartphone e cuffie al collo, operaio con gilet catarifrangente, sportivo con fascetta, papà, chef).
* **Regole Auree estratte**:
  * **Ponte tra Anime e Realtà Urbana**: Fornisce il vocabolario visivo per gli abiti moderni (**streetwear casual per Alfredo**, **tailleur per la Dirigente e la Rettrice**, **gilet/divise per i Vigili e la Polizia 1★–3★**, **smartphone e cuffie per i passanti e le Studentesse**).
  * **Nota di Fusione Stilistica**: Prendiamo i **dettagli d'abbigliamento moderno** di `peak-npcref1.png` e li caliamo nelle **proporzioni Chibi super-espressive e rifinite** di `peak-npcref4/5/6`!

#### 4. `peak-envref7.jpg` & `peak-envref4.jpg` — *Direzione Artistica dell'Ambiente e delle Pensiline*
* **Cosa mostrano**:
  * [`peak-envref7.jpg`](./ref/peak-envref7.jpg): Facciata di un negozio/konbini giapponese in pixel art dettagliatissima: tendone a righe rosa corallo e crema, **distributore automatico rosso di bevande** (`Distributore Neon` in-game!), insegna giapponese blu notte, unità esterna del condizionatore avvolta dall'edera, scatoloni con gattino nero, mascotte rosa.
  * [`peak-envref4.jpg`](./ref/peak-envref4.jpg): Metropoli isometrica iper-popolata e coloratissima (*"HIPSTER'S", "BOOKS", "WINERY", "DANCEDANCE", "ARTMART"*), piena di auto colorate, ambulanze e micro-scene urbane su ogni marciapiede e tetto.
* **Svolta Estetica Fondamentale per `girlbusAG`**:
  * **Addio al "Grimdark Cupo", Benvenuto al "Vibrant Pop-Urban"**: L'arena di `girlbusAG` **NON è un vicolo buio, triste e monocromatico**. È un **quartiere universitario/commerciale vibrante, saturo e vivo** (stile Tokyo/Roma Pop-University), pieno di **tende a righe, distributori automatici colorati, vegetazione urbana che spunta dal cemento, insegne ironiche e traffico folle**!
  * **Contrasto Tattico per il Bullet-Hell**:
    * **Pavimento (Asfalto & Marciapiedi)**: Toni medi caldi ma desaturati (grigio-malva asfalto `#585568`, beige pietra marciapiede `#C8BFA8`) per non stancare la vista.
    * **Edifici, Pensiline & Chioschi (`peak-envref7`)**: Colori ricchi, caldi e saturi (rosa corallo, crema, rosso vending machine, verde foglia, blu lavagna).
    * **Proiettili, Raggi Fermate & Ologrammi Minigiochi**: Colori puri ad altissima luminanza (Ciano Elettrico `#00F5D4`, Rosa Shocking `#FF2A85`, Giallo Laser `#FFE600`) con contorno scuro/bianco per bucare lo schermo!

---

### 🏙️ C. Riferimenti Secondari (`envref1, 2, 3, 5, 6` & `npcref2, 3`) — *Layout Camera & Leggibilità Top-Down*

| File | Cosa Aggiunge alla Produzione | Cosa Scartiamo / Adattiamo |
| :--- | :--- | :--- |
| **[`envref3.jpg`](./ref/envref3.jpg)** | **Oro colato per il Layout Stradale**: mostra la **piazzola gialla `BUS` sull'asfalto**, la **pensilina dell'autobus** sul marciapiede, le strisce pedonali, i semafori e i parcheggi in **prospettiva Top-Down 3/4 ortogonale** (perfetta per muoversi a 360° e leggere i cerchi delle fermate). | Lo shading di `envref3` è un po' piatto/slavato; useremo la **palette calda e croccante di `peak-envref7`**. |
| **[`envref2.jpg`](./ref/envref2.jpg)** & **[`envref5.jpg`](./ref/envref5.jpg)** | Ottima scala tra personaggi chibi, **distributori automatici (`envref5`)**, caffetterie (*"COFFEE DAILY"*), negozi tech (*"TOP TOWER"*) e poliziotti davanti agli edifici. | I personaggi di `envref2/5` sono troppo piccoli/semplici (`16×24`); noi useremo il dettaglio superiore di **`peak-npcref4/5/6` (`32×32`)**. |
| **[`envref1.png`](./ref/envref1.png)** | Struttura chiara degli incroci a 4 vie, marciapiedi, lampioni e parco urbano con alberi arrotondati. | Palette troppo "retro 8-bit" satura sul verde prato; preferiamo il verde foglia morbido di `peak-envref7`. |
| **[`envref6.png`](./ref/envref6.png)** | Atmosfera arcade (*"PIXELADE ARCADE"*), cartello stradale **`STOP`** (ottimo riferimento per i segnali delle ZTL!) e pose dinamiche di combattimento. | La vista è laterale da picchiaduro a scorrimento; il nostro gioco usa la camera **Top-Down 3/4**. |
| **[`npcref2.png`](./ref/npcref2.png)** & **[`npcref3.jpg`](./ref/npcref3.jpg)** | Confermano la pulizia dell'outline `1px` scuro e la leggibilità delle pose di 3/4 (`@LANCEKIM0806`). | Per le ragazze, gli occhi e i dettagli dei capelli seguiranno la ricchezza di **`peak-npcref4/5`**. |

---

## 2. Specifiche Tecniche degli Sprite (Standard per Produzione & Subagenti)

Per garantire che i **3 sprite master** creati ora e tutti gli **sprite secondari + animazioni** delegati ai subagenti *Hydrogen (Medium/High)* siano **100% coerenti al pixel**, fissiamo queste regole matematiche:

### 2.1 Griglia, Risoluzione e Proporzioni (`32×32 px` Standard Canvas)
* **Canvas Base Personaggi (Alfredo, Ragazze, Polizia)**: **`32×32 pixel`** (con margine fino a **`36×36 px`** per Boss o cappelli/code voluminose).
  * **Altezza Effettiva Personaggio**: **`28–30 px`** centrati in basso (lasciando `2–4 px` sopra la testa per il rimbalzo dell'animazione di corsa).
  * **Testa (inclusi capelli)**: **`16–17 px` di altezza × `16–20 px` di larghezza** (~58% dell'altezza totale).
  * **Busto + Braccia**: **`7–8 px` di altezza**.
  * **Gambe + Scarpe**: **`4–5 px` di altezza** (piedini compatti `3×2` o `3×3` px che si alternano chiaramente nei frame di corsa).
  * **Ombra a Terra**: Ellisse semitrasparente (`14×4 px`, `#1A1025` al `45%` alpha) sotto i piedi per ancorare lo sprite all'asfalto.
* **Canvas Pensiline / Distributori (Fermate dell'Autobus & Chioschi)**: **`48×48 pixel`** (così che un personaggio `32×32` ci stia comodamente sotto/accanto!).
* **Canvas Autobus Neon**: **`96×48 pixel`** (orizzontale) / **`48×96 pixel`** (verticale).

### 2.2 Regole di Disegno Pixel-Perfect (Vincoli per Subagenti)
1. **Zero "Mixels"**: Ogni singolo pixel è `1×1` nella griglia nativa `32×32`. Per le anteprime utente esportiamo sempre sia il file **`1x` nativo (`32×32`)** sia il file **`10x` Nearest-Neighbor (`320×320`)** nitidissimo.
2. **Outline Selettivo `1px` (`#1C1124`)**:
   * Contorno esterno continuo di `1px` (senza "doppi angoli" spessi a scaletta — *jaggies*).
   * Niente outline nero puro `#000000`: usiamo **Deep Aubergine (`#1C1124`)**, lo stesso tono caldo-scuro visibile in `peak-npcref4`, `peak-npcref5` e `peak-envref7`.
3. **Shading a 3 Livelli (Cel-Pixel Shading)**:
   * Luce proveniente dall'**Alto-Sinistra** (`Top-Left`).
   * Ogni superficie (pelle, vestiti, capelli, strutture) usa esattamente **Base + Ombra (`-20% lum, +shift verso il viola/blu`) + Highlight (`+20% lum, +shift verso il giallo/crema`)**.

### 2.3 Master Palette Unificata (`girlbusAG Pop-Urban Palette`)

| Gruppo | Ruolo | Hex Base | Hex Ombra | Hex Luce / Accento | Riferimento `ref/` |
| :--- | :--- | :---: | :---: | :---: | :--- |
| **Outline & Ombre** | Contorno Universale `1px` | `#1C1124` | `#120A18` | `#352342` (Outline interno) | `peak-npcref4 / 5 / 7` |
| **Pelle (Skin)** | Incarnato Caldo Chibi | `#FFE0C8` | `#F0B89A` | `#FFF2E6` (Blush: `#FF8FA3`) | `peak-npcref4 / 5 / 6` |
| **Alfredo (Capelli)** | Castano Scuro / Nero Corto | `#282024` | `#181216` | `#45373E` (Sopracciglia: `#1C1124`) | `mainref-alfredo1 / 2` |
| **Alfredo (Outfit)** | Streetwear Jacket & Tee | `#2E5A88` | `#1D3B5C` | `#4B7FB3` (Tee: `#F5F7FA`) | `peak-npcref1 / envref5` |
| **Alfredo (Jeans/Shoes)**| Pantaloni Scuri & Sneakers | `#333947` | `#222630` | `#E63946` (Sneaker Accent) | `peak-npcref1` |
| **Archetipo Kawaii** | Rosa Sakura & Crema | `#FF6B9D` | `#D93870` | `#FFA8C5` (Accento: `#FFF0F5`) | `peak-npcref4 / envref7` |
| **Archetipo Emo** | Viola Grunge & Nero | `#6D3B8E` | `#43205C` | `#9D65C2` (Accento: `#E05288`) | `peak-npcref4` |
| **Archetipo Studentessa**| Blu Campus & Caramello | `#3B6998` | `#244366` | `#D98A4E` (Cardigan/Capelli) | `peak-npcref1 / 4` |
| **Archetipo Cyber-Gyaru**| Biondo Pop & Ciano 5G | `#FFB84D` | `#E08924` | `#00F0FF` (Ologramma Neon) | `peak-npcref4` |
| **Ambiente & Strada** | Asfalto, Marciapiede, Konbini | `#5A566B` | `#3F3B4D` | `#D8CEBE` (Marciapiede caldo) | `peak-envref7 / envref3` |
