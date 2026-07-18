# oauth-login — Design Index

Each Decision maps to its `DNN.md` file; every minted `R-nnnn-nnnn` id maps to
the Decision that owns it and the file it lives in. Id lookup is a grep against
this index:

    grep -n R-nnnn-nnnn project/design/INDEX.md

Regenerate this file whenever a Decision is added or its Verification ids
change.

## Decisions

- **D1** → `D01.md` — Package decomposition and seams — none; structural.
- **D2** → `D02.md` — PKCE and authorize-URL construction — R-0EF0-5TL8,
  R-0FMW-JLBX, R-0GUS-XD2M, R-0I2P-B4TB, R-0JAL-OWK0, R-0KII-2OAP,
  R-0LQE-GG1E, R-0MYA-U7S3, R-0O67-7ZIS, R-0PE3-LR9H.
- **D3** → `D03.md` — Token exchange and verbatim passthrough — R-0QLZ-ZJ06,
  R-0RTW-DAQV, R-0U9P-4U89, R-0VHL-ILYY, R-0WPH-WDPN, R-0XXE-A5GC,
  R-0Z5A-NX71, R-10D7-1OXQ.
- **D4** → `D04.md` — The loopback callback listener — R-11L3-FGOF,
  R-12SZ-T8F4, R-140W-705T, R-158S-KRWI, R-16GO-YJN7, R-18WH-Q34L,
  R-1A4E-3UVA, R-1BCA-HMLZ, R-1CK6-VECO, R-1DS3-963D, R-1EZZ-MXU2,
  R-1G7W-0PKR.
- **D5** → `D05.md` — The browser launch seam — R-1HFS-EHBG, R-1INO-S925,
  R-1JVL-60SU, R-1L3H-JSJJ, R-1MBD-XKA8.
- **D6** → `D06.md` — CLI surface, validation, and the composition root —
  R-1NJA-BC0X, R-1PZ3-2VIB, R-1R6Z-GN90, R-1SEV-UEZP, R-1TMS-86QE,
  R-1UUO-LYH3, R-1W2K-ZQ7S, R-1XAH-DHYH, R-1YID-R9P6, R-1ZQA-51FV,
  R-20Y6-IT6K.
- **D7** → `D07.md` — Live verification against a real provider — R-2262-WKX9.
- **D8** → `D08.md` — Build and install tooling — none; structural.

## Verification ids → Decision

| id | Decision | file |
|---|---|---|
| R-0EF0-5TL8 | D2 | `D02.md` |
| R-0FMW-JLBX | D2 | `D02.md` |
| R-0GUS-XD2M | D2 | `D02.md` |
| R-0I2P-B4TB | D2 | `D02.md` |
| R-0JAL-OWK0 | D2 | `D02.md` |
| R-0KII-2OAP | D2 | `D02.md` |
| R-0LQE-GG1E | D2 | `D02.md` |
| R-0MYA-U7S3 | D2 | `D02.md` |
| R-0O67-7ZIS | D2 | `D02.md` |
| R-0PE3-LR9H | D2 | `D02.md` |
| R-0QLZ-ZJ06 | D3 | `D03.md` |
| R-0RTW-DAQV | D3 | `D03.md` |
| R-0U9P-4U89 | D3 | `D03.md` |
| R-0VHL-ILYY | D3 | `D03.md` |
| R-0WPH-WDPN | D3 | `D03.md` |
| R-0XXE-A5GC | D3 | `D03.md` |
| R-0Z5A-NX71 | D3 | `D03.md` |
| R-10D7-1OXQ | D3 | `D03.md` |
| R-11L3-FGOF | D4 | `D04.md` |
| R-12SZ-T8F4 | D4 | `D04.md` |
| R-140W-705T | D4 | `D04.md` |
| R-158S-KRWI | D4 | `D04.md` |
| R-16GO-YJN7 | D4 | `D04.md` |
| R-18WH-Q34L | D4 | `D04.md` |
| R-1A4E-3UVA | D4 | `D04.md` |
| R-1BCA-HMLZ | D4 | `D04.md` |
| R-1CK6-VECO | D4 | `D04.md` |
| R-1DS3-963D | D4 | `D04.md` |
| R-1EZZ-MXU2 | D4 | `D04.md` |
| R-1G7W-0PKR | D4 | `D04.md` |
| R-1HFS-EHBG | D5 | `D05.md` |
| R-1INO-S925 | D5 | `D05.md` |
| R-1JVL-60SU | D5 | `D05.md` |
| R-1L3H-JSJJ | D5 | `D05.md` |
| R-1MBD-XKA8 | D5 | `D05.md` |
| R-1NJA-BC0X | D6 | `D06.md` |
| R-1PZ3-2VIB | D6 | `D06.md` |
| R-1R6Z-GN90 | D6 | `D06.md` |
| R-1SEV-UEZP | D6 | `D06.md` |
| R-1TMS-86QE | D6 | `D06.md` |
| R-1UUO-LYH3 | D6 | `D06.md` |
| R-1W2K-ZQ7S | D6 | `D06.md` |
| R-1XAH-DHYH | D6 | `D06.md` |
| R-1YID-R9P6 | D6 | `D06.md` |
| R-1ZQA-51FV | D6 | `D06.md` |
| R-20Y6-IT6K | D6 | `D06.md` |
| R-2262-WKX9 | D7 | `D07.md` |
</content>
