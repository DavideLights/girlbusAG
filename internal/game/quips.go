package game

import "fmt"

// BossIntroQuip returns a short Italian arcade quip when a Gatekeeper Boss enters the arena.
func BossIntroQuip(kind BossKind) string {
	switch kind {
	case BossBestFriend:
		return "Ehi! Lei stasera non sale su nessun bus!"
	case BossLover:
		return "Trova quella vera tra i miei ologrammi!"
	case BossProfessor:
		return "Esame a sorpresa: zero errori in combo!"
	case BossExecutive:
		return "Serve il timbro delle mie stagiste, caro!"
	case BossRector:
		return "Qui comando io: ZTL marziale attiva!"
	case BossMom:
		return "Alfredo! Torna subito a casa, disgraziato!"
	default:
		return "Pronto al capolinea? Vediamo chi frena!"
	}
}

// BossBreakQuip returns a short Italian quip when a Boss shield breaks or when the Boss is defeated.
func BossBreakQuip(kind BossKind, shieldsLeft int) string {
	if shieldsLeft <= 0 {
		switch kind {
		case BossBestFriend:
			return "Va bene, hai vinto... portala al capolinea!"
		case BossLover:
			return "Okay, hai piu rizz dei miei riflessi!"
		case BossProfessor:
			return "Trenta e lode! Promosso a pieni voti!"
		case BossExecutive:
			return "Sei assunto... guidi tu la linea notte!"
		case BossRector:
			return "Laurea honoris causa in guida spericolata!"
		case BossMom:
			return "Va bene, guida piano e mettiti la sciarpa!"
		default:
			return "Via libera! Prossima fermata!"
		}
	}

	switch kind {
	case BossBestFriend:
		return "Tsk, colpo di fortuna! Non mollo mica!"
	case BossLover:
		return "Ahia! Era quella vera? Riproviamoci!"
	case BossProfessor:
		return "Sufficiente... ma il prossimo giro e duro!"
	case BossExecutive:
		return "Pratica approvata! Ora passa alla prossima!"
	case BossRector:
		return "Impertinente! Alza la guardia, matricola!"
	case BossMom:
		return "Aspetta che prendo l'altra ciabatta!"
	default:
		return "Scudo incrinato! Continua cosi!"
	}
}

// SynergyQuip returns a short Italian arcade quip when an archetype Synergy activates.
func SynergyQuip(arch Archetype) string {
	switch arch {
	case ArchEmo:
		return "Anima nera sintonizzata: SOULMATE!"
	case ArchStudent:
		return "Appunti perfetti: HONOR ROLL attiva!"
	case ArchKawaii:
		return "Occhi a cuore: MOE CRITICAL x2!"
	case ArchGyaru:
		return "In tendenza stanotte: VIRAL #1!"
	default:
		return "Sintonizzati sulla stessa pensilina!"
	}
}

// EndRunQuip returns a tailored Italian summary quip for the Victory / Game Over screen.
func EndRunQuip(victory, bossRush bool, peakWanted, bossesDefeated, maxStreak int) string {
	if victory {
		if bossRush {
			return "BOSS RUSH DOMINATA! SEI LA LEGGENDA!"
		}
		if peakWanted >= 5 {
			return "VITTORIA CON 5* WANTED! RE DELLA NOTTE!"
		}
		if maxStreak >= 5 {
			return "CORSA PERFETTA! RUSH HOUR INARRESTABILE!"
		}
		return "TUTTE A BORDO! TURNO DI NOTTE SUPERATO!"
	}

	if bossRush {
		if bossesDefeated >= 4 {
			return "QUASI IN CIMA ALLA BOSS RUSH! RIPROVA!"
		}
		return "LA BOSS RUSH NON PERDONA! TORNA IN PISTA!"
	}
	if peakWanted >= 4 {
		return "TROPPO HEAT POLIZIA! USA LE TANGENTI!"
	}
	if bossesDefeated == 0 {
		return "OCCHIO AI PROIETTILI: USA DASH E CHRONOS!"
	}
	return "BUONA CORSA! IL PROSSIMO TURNO E TUO!"
}

// FormatRunSeed formats a run RNG seed into a compact 6-digit hex tag (#XXXXXX).
func FormatRunSeed(seed int64) string {
	return fmt.Sprintf("#%06X", uint32(seed)&0xFFFFFF)
}

// ShowQuip activates a timed micro-dialogue banner in the simulation state.
func (w *World) ShowQuip(speaker, text string, accent int, duration float32) {
	if text == "" {
		return
	}
	if duration <= 0 {
		duration = 3.0
	}
	w.QuipSpeaker = speaker
	w.QuipText = text
	w.QuipAccent = accent
	w.QuipTimer = duration
	w.QuipMaxTimer = duration
}
