package main

// The array index is the ticket number minus one. Keep this order unchanged
// after journal.xlsx has been used, so previous ticket numbers keep their topic.
var questions = [ticketCount]string{
	"Ce este un sistem informațional și care sunt componentele sale principale?",
	"Cum identificați părțile interesate și obiectivele unui sistem informațional nou?",
	"Care este diferența dintre cerințele funcționale și cele nefuncționale? Dați exemple.",
	"Comparați modelul în cascadă cu modelul incremental al ciclului de viață.",
	"Care sunt etapele principale ale ciclului de viață al unui sistem informațional?",
	"Ce se analizează în studiul preliminar al obiectului de informatizare?",
	"Ce rezultate trebuie să producă etapa de analiză detaliată a sistemului existent?",
	"Ce reprezintă diagrama fluxurilor de date (DFD) și ce elemente conține?",
	"Cum se realizează descompunerea unui proces DFD fără a pierde fluxurile externe?",
	"Ce reprezintă intrările, ieșirile, mecanismele și controalele în IDEF0?",
	"Pentru ce se utilizează IDEF3 în modelarea proceselor?",
	"Ce rol are IDEF1X în proiectarea modelului de date?",
	"Cum identificați entitățile, relațiile și cardinalitățile într-un model de date?",
	"Ce descrie o diagramă UML de cazuri de utilizare și care este rolul actorilor?",
	"Comparați diagrama de clase UML cu diagrama de secvență UML.",
	"Ce avantaje și limite are arhitectura client–server pentru un sistem informațional?",
	"Cum ajută instrumentele CASE la analiza și proiectarea sistemelor informaționale?",
	"După ce criterii comparați două variante de proiectare a unui sistem informațional?",
	"Ce categorii de procese ale ciclului de viață sunt descrise de ISO/IEC 12207?",
	"Cum verificați că sistemul implementat respectă cerințele documentate?",
}
