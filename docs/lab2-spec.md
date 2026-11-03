# Lab 2 — implementation record

This branch implements the user's Lab 2 DOCX assignment as a continuation of the Go console ticket generator. The subject is **Proiectarea Sistemelor Informaționale**. Demo inputs use fictional students and 20 locally authored tickets, each with three questions.

The GUI loads groups from workbook sheet names, lists students from columns A and B, and enables generation only after both selections. Every generation appends to `results.xlsx`. A repeat selection reuses the earliest recorded ticket for the exact group, last name, and first name, while still adding a new result row marked `да`. A new selection uses a cryptographically random valid ticket ID. Saving completes before the result window opens. ESC closes the result window and returns to the selector.

The implementation reports missing or invalid input files, skips incomplete ticket blocks, and offers a retry for failed result saves. The same ticket remains pending during a retry. The generated Windows artifact contains the GUI executable and both demo inputs, so the target computer needs no Go installation.

OpenAI Codex authored this implementation. The assignment mentioned Copilot Chat as a development aid; the user explicitly accepted Codex and a visible GitHub history instead. This record describes the implementation and does not claim use of Copilot.
