# Bookshelf: Briefing für Hi-Fi-Mockups

Abgeleiteter Arbeitsstand für eine **Designexploration** aus dem Beispiel-Spec-Set, keine freigegebene UX-Spezifikation. Die drei Entwürfe sind Vorschläge für dieselbe Aufgabe, keine freigegebenen Darstellungsmodi. Es wurden keine Bilder erzeugt und keine Inhalte an einen externen Dienst übertragen.

## Verwendung

Den folgenden eingezäunten Prompt vollständig in einen geeigneten Design-Agenten kopieren. Vor externer Nutzung die Freigabe der enthaltenen Informationen für das Ziel prüfen; keine privaten Spezifikationen und keine echten Nutzerdaten hochladen. Quellen, offene Entscheidungen und Prüfliste unterhalb des Prompts bleiben beim lokalen Auftrag.

```text
Du gestaltest konkrete, lesbare Hi-Fi-Mockups für „Bookshelf“.
Erzeuge Bilder, nicht nur eine Beschreibung, wie Bilder aussehen könnten.
Falls du keine Bildgenerierung, aber HTML/SVG und Browser-Rendering beherrschst,
liefere renderbare Quelldateien und daraus exportierte Bilder. Falls du keine
Bilder exportieren kannst, sage das ausdrücklich; erfinde keine Dateilinks.

AUFTRAG UND STATUS
Entwirf drei vergleichbare visuelle Konzepte für denselben Ablauf: ein Buch
zum Regal hinzufügen. Dies ist eine Exploration, keine Freigabe zur Änderung
des Produkts. Die Regeln von Release R0.1 unten sind verbindlich. Die drei
visuellen Richtungen und alle ergänzenden Beschriftungen sind Vorschläge.
Erfinde keine Funktionen, Rollen oder Nachweise.

NUTZER UND ZWECK
Bookshelf hält eine Leseliste fest. Nutzer: eine Person mit etwa hundert
Büchern, die in Schüben liest, meist am Smartphone, oft einhändig, bei
wechselndem Licht. Das Ziel ist, ein Buch in unter zehn Sekunden hinzuzufügen;
das ist eine zu prüfende Hypothese, kein gemessenes Ergebnis und kein
Werbeversprechen.

MARKE — VORLÄUFIGER VORSCHLAG
Zeige „Bookshelf“ als schlichte Wortmarke. Kein erfundenes Logo. Die
Markenfrage ist offen; kennzeichne das nur in der Bildunterschrift.

GRENZE VON R0.1
Nur Regalliste und Formular. Keine Suche, keine Schlagworte, kein Konto,
kein Löschen. „Als gelesen markieren“ gehört zu Release R0.2 und fehlt in
diesen Bildern.

GLEICHER ABLAUF, GLEICHE DATEN
Regal ohne Bücher → „Buch hinzufügen“ → Formular mit Titel und Autor →
Speichern → Regal mit einem Buch.
Das Formular zeigt genau zwei Felder, Titel und Autor, und den Knopf
„Buch hinzufügen“. Kein drittes Pflichtfeld. Ein abgelehnter Titel zeigt die
Fehlermeldung unter dem Feld; das Regal bleibt unverändert.
Daten, rein synthetisch: Titel „Beispielbuch“, Autor „Beispielautor“.

VISUELLE BASIS AUS DER UX
Oberfläche #FAF7F2; Text #1F1B16; Akzent #2F6F5E; Rahmen #D8D2C8.
Systemschrift, Fließtext 18 px; Touch-Ziele mindestens 48 px.
Kontraste prüfen, nicht behaupten. Kein winziger grauer Text.

DREI VISUELLE KONZEPTE — B UND C SIND TOKEN-VORSCHLÄGE
A „Papier“: Ausgangspunkt mit den Tokens oben.
B „Abend“: Vorschlag für geringe Blendung. Hintergrund #16191C, Oberfläche
  #20252A, Text #ECE8E0, Akzent #6FB8A4.
C „Kontrast“: Vorschlag für maximale Erkennbarkeit. Hintergrund #FFFFFF,
  Text #000000, Rahmen #000000, Aktionsfläche #FFD84D mit schwarzer
  Beschriftung. Status nicht nur durch Farbe unterscheiden.
Prüfe Kontraste auch für B und C. Keines der Konzepte behauptet
WCAG-Konformität oder eine nachgewiesene Präferenz. Alle drei zeigen dieselben
Inhalte, Aktionen und Feldreihenfolgen. Die Variantenauswahl gehört in einen
äußeren Vorschau-Rahmen, nicht in das Formular.

DEUTSCHE UI-TEXTE — QUELLENTREU VERWENDEN
W-MC-01: „Buch hinzufügen“
W-MC-02: „Zum Regal hinzugefügt“
W-MC-03: „Der Titel fehlt oder ist zu lang.“
W-MC-05: „Dein Regal ist noch leer. Füge das erste Buch hinzu.“
Kein „Book“, „Add“ oder „Submit“ im UI. Umlaute korrekt setzen; keine
abgeschnittenen Wörter, Pseudobuchstaben oder unlesbaren Bildtexte.

ERGÄNZENDE TEXTE — VORSCHLÄGE, NICHT BEREITS SPEZIFIZIERT
P-COPY-01: „Titel“ und „Autor“ als Feldbeschriftungen.
P-COPY-02: „Zurück zum Regal“ als Beschriftung der Rücknavigation.
Nutze sie in der Exploration und liste sie im Begleittext als freigabepflichtig.

BILDAUFTRAG
1. Drei einzelne PNGs mit dem ausgefüllten Formular: shelf-form-A.png,
   shelf-form-B.png, shelf-form-C.png, jeweils 390 × 844 Pixel.
2. shelf-variants.png: daraus eine Vergleichstafel nebeneinander. Jede Ansicht
   bleibt 390 × 844; Beschriftungen in zusätzlichem Außenrand, nicht über dem UI.
3. Für Konzept A, ohne es dadurch als Gewinner festzulegen, jeweils 390 × 844:
   shelf-flow-01.png: leeres Regal mit W-MC-05;
   shelf-flow-02.png: Formular mit abgelehntem Titel und W-MC-03;
   shelf-flow-03.png: Regal mit einem Buch und dem Hinweis W-MC-02.
Erzeuge keine zusätzlichen Screens für jede denkbare Kombination.

LIEFERE ZUM SCHLUSS
Die tatsächlich erzeugten Dateien mit Abmessungen und, falls vorhanden,
editierbaren Quellen; kurze Begründung je Konzept mit Nutzerbezug;
Abweichungen von den Tokens und Liste der vorgeschlagenen Texte; offene
Entscheidungen, technische Einschränkungen und erkennbare Textfehler.
Kennzeichne alle Bilder außerhalb des UI als „Designentwurf · Beispieldaten“.
Nenne nichts „freigegeben“, „barrierefrei geprüft“ oder „mit Nutzern
validiert“, solange der jeweilige Nachweis fehlt.
```

## Lokaler Quellennachweis

Die Quellen sind Teil dieses Repositories; ihre Revision ist die des Checkouts. Die Hashes identifizieren exakt den gelesenen Inhalt.

| Quelle | SHA-256 | Verwendete Abschnitte |
|---|---|---|
| [PS](lead/BOOKSHELF_PS.md) | `c13cfc74eb0a0c88d9a31f27573d1a19d9f53ee8f2b8cb20006ea606bfa08c82` | §2 Persona und JTBD; §3 Erfolgshypothese |
| [PRD](pm/BOOKSHELF_PRD.md) | `39e5d46771ab4a9558a96257e767fce80fef2e759dae65cdf099048441c8e268` | §1 Scope; AC-SHELF-01, AC-SHELF-03 |
| [UX](pm/ux/BOOKSHELF_UX.md) | `cc3679c7da3aca4458645f95f7cd686672ac6403704b89b565c689a449d1ead9` | §2 Tokens; §5 W-MC-01 bis W-MC-03 und W-MC-05; UXB-01.2 |
| [BACKLOG](pm/BOOKSHELF_BACKLOG.md) | `048b01b099f0e2e2dbcab27985bd4d96bb687fa8cfc8d655b7e2a4dc3ceceb7b` | BK-001 (R0.1); BK-002 (R0.2) nur als Grenze |

Der Auftrag des Menschen liefert die Anzahl drei und den Wunsch nach deutschem Text. Konzeptnamen, die Tokenwerte von B und C, die provisorische Wortmarke, die Bildmaße und P-COPY-01/02 sind **neue Vorschläge**, keine Quellenzitate.

### Offene Entscheidungen und Lücken

| ID | Befund / Wirkung | Zuständig | Exploration / Umsetzung |
|---|---|---|---|
| G-01 | Markenauftritt offen; die Wortmarke im Prompt ist nur ein Vorschlag | Produktverantwortung | Exploration möglich; vor Freigabe entscheiden |
| G-02 | Die UX nennt ein Farbschema, keine Varianten; B und C sind neue Token-Vorschläge | UX | Vergleich möglich; Auswahl vor Umsetzung |
| G-03 | Feldbeschriftungen und Rücknavigation (P-COPY-01/02) sind nicht spezifiziert | UX | explorieren; Text vor Umsetzung in der UX festlegen |
| G-04 | Die Zehn-Sekunden-Hypothese ist nicht beobachtet | Produktverantwortung | Bildexploration möglich; Feldnachweis bleibt offen |

### Prüfliste für zurückgegebene Assets

- [ ] Dateien wirklich vorhanden; Einzelbilder 390 × 844, Vergleichstafel ohne Verkleinerung der Einzelansichten.
- [ ] Alle drei Konzepte zeigen denselben Zustand und dieselben Daten; Unterschiede sind benannt, keine neuen Funktionen.
- [ ] „Zum Regal hinzugefügt“, „Der Titel fehlt oder ist zu lang.“ und alle verwendeten Quellentexte korrekt, lesbar und nicht abgeschnitten.
- [ ] Genau zwei Felder; kein Konto, keine Suche, kein Löschen; „Als gelesen markieren“ fehlt (R0.2).
- [ ] Fehler und Erfolg eindeutig getrennt; die Eingabe nicht als gespeichert dargestellt, wenn sie abgelehnt wurde.
- [ ] Farben, Schrift- und Touchgrößen und Kontraste tatsächlich geprüft; Abweichungen von den Tokens benannt.
- [ ] Markenentscheidung, Textvorschläge und Nutzerprüfung separat freigegeben bzw. weiterhin offen.

Die technische Briefing-Prüfung ersetzt keine dieser Bild- und Nutzerprüfungen. Nach relevanten Spec-Änderungen neu extrahieren; Korrekturen an Anforderungen gehören in deren Eigentümerdateien, nicht in diesen Snapshot.
