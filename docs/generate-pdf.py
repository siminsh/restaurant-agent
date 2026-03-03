#!/usr/bin/env python3
"""Convert project-report.md to a styled PDF using fpdf2 with Unicode TTF fonts."""

import os
import re

from fpdf import FPDF

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
MD_PATH = os.path.join(SCRIPT_DIR, "project-report.md")
PDF_PATH = os.path.join(SCRIPT_DIR, "project-report.pdf")

# macOS TTF font paths
FONT_DIR = "/System/Library/Fonts/Supplemental"
ARIAL = os.path.join(FONT_DIR, "Arial.ttf")
ARIAL_BOLD = os.path.join(FONT_DIR, "Arial Bold.ttf")
ARIAL_ITALIC = os.path.join(FONT_DIR, "Arial Italic.ttf")
ARIAL_BOLD_ITALIC = os.path.join(FONT_DIR, "Arial Bold Italic.ttf")
COURIER = os.path.join(FONT_DIR, "Courier New.ttf")
COURIER_BOLD = os.path.join(FONT_DIR, "Courier New Bold.ttf")

# Colors
ORANGE = (255, 107, 53)
BLUE = (0, 78, 137)
TEAL = (46, 196, 182)
DARK = (26, 26, 26)
GRAY = (100, 100, 100)
CODE_BG = (42, 42, 42)
CODE_FG = (212, 212, 212)
WHITE = (255, 255, 255)
TABLE_HEADER_BG = (0, 78, 137)
TABLE_ALT_BG = (248, 249, 250)

# Font family names
F_BODY = "Arial"
F_CODE = "CourierNew"


class ProjectPDF(FPDF):
    def __init__(self):
        super().__init__()
        self.set_auto_page_break(auto=True, margin=20)
        # Register Unicode TTF fonts
        self.add_font(F_BODY, "", ARIAL, uni=True)
        self.add_font(F_BODY, "B", ARIAL_BOLD, uni=True)
        self.add_font(F_BODY, "I", ARIAL_ITALIC, uni=True)
        self.add_font(F_BODY, "BI", ARIAL_BOLD_ITALIC, uni=True)
        self.add_font(F_CODE, "", COURIER, uni=True)
        self.add_font(F_CODE, "B", COURIER_BOLD, uni=True)

    def header(self):
        if self.page_no() == 1:
            return
        self.set_font(F_BODY, "I", 8)
        self.set_text_color(*GRAY)
        self.cell(0, 8, "De Gouden Lepel \u2014 Restaurant Operations Agent", align="L")
        self.ln(4)

    def footer(self):
        self.set_y(-15)
        self.set_font(F_BODY, "I", 8)
        self.set_text_color(*GRAY)
        self.cell(0, 10, f"Page {self.page_no()}/{{nb}}", align="C")

    def write_title(self, text):
        self.set_font(F_BODY, "B", 24)
        self.set_text_color(*ORANGE)
        self.cell(0, 14, text, new_x="LMARGIN", new_y="NEXT")
        self.set_draw_color(*ORANGE)
        self.set_line_width(0.8)
        self.line(self.l_margin, self.get_y(), self.w - self.r_margin, self.get_y())
        self.ln(6)

    def write_subtitle(self, text):
        self.set_font(F_BODY, "I", 12)
        self.set_text_color(*GRAY)
        self.cell(0, 8, text, new_x="LMARGIN", new_y="NEXT")
        self.ln(4)

    def write_h2(self, text):
        self.ln(6)
        if self.get_y() > 250:
            self.add_page()
        self.set_font(F_BODY, "B", 16)
        self.set_text_color(*BLUE)
        self.cell(0, 10, text, new_x="LMARGIN", new_y="NEXT")
        self.set_draw_color(220, 220, 220)
        self.set_line_width(0.3)
        self.line(self.l_margin, self.get_y(), self.w - self.r_margin, self.get_y())
        self.ln(4)

    def write_h3(self, text):
        self.ln(4)
        if self.get_y() > 260:
            self.add_page()
        self.set_font(F_BODY, "B", 13)
        self.set_text_color(*TEAL)
        self.cell(0, 9, text, new_x="LMARGIN", new_y="NEXT")
        self.ln(2)

    def write_paragraph(self, text):
        self.set_font(F_BODY, "", 10)
        self.set_text_color(*DARK)
        self._write_rich_text(text)
        self.ln(4)

    def _write_rich_text(self, text):
        """Write text with inline bold, code, and italic support."""
        parts = re.split(r'(\*\*.*?\*\*|`[^`]+`|\*[^*]+\*)', text)
        self.set_font(F_BODY, "", 10)
        self.set_text_color(*DARK)
        for part in parts:
            if part.startswith("**") and part.endswith("**"):
                self.set_font(F_BODY, "B", 10)
                self.write(5.5, part[2:-2])
                self.set_font(F_BODY, "", 10)
            elif part.startswith("`") and part.endswith("`"):
                self.set_font(F_CODE, "", 9)
                self.set_text_color(180, 40, 40)
                self.write(5.5, part[1:-1])
                self.set_font(F_BODY, "", 10)
                self.set_text_color(*DARK)
            elif part.startswith("*") and part.endswith("*") and len(part) > 2:
                self.set_font(F_BODY, "I", 10)
                self.write(5.5, part[1:-1])
                self.set_font(F_BODY, "", 10)
            else:
                self.write(5.5, part)
        self.ln()

    def write_blockquote(self, text):
        y = self.get_y()
        # Calculate needed height
        self.set_font(F_BODY, "I", 10)
        cell_w = self.w - self.l_margin - self.r_margin - 12
        n_lines = max(1, len(text) * self.get_string_width("x") / cell_w + 1)
        block_h = max(12, n_lines * 5.5 + 4)
        self.set_fill_color(255, 248, 244)
        self.rect(self.l_margin, y, self.w - self.l_margin - self.r_margin, block_h, "F")
        self.set_draw_color(*ORANGE)
        self.set_line_width(1)
        self.line(self.l_margin, y, self.l_margin, y + block_h)
        self.set_xy(self.l_margin + 6, y + 2)
        self.set_text_color(80, 80, 80)
        self.multi_cell(cell_w, 5.5, text)
        self.set_y(y + block_h + 2)
        self.ln(4)

    def write_code_block(self, lines):
        self.ln(2)
        y_start = self.get_y()
        line_h = 4.2
        n_lines = len(lines)
        block_h = n_lines * line_h + 10

        if y_start + block_h > self.h - 20:
            if block_h < self.h - 40:
                self.add_page()
                y_start = self.get_y()
            # If block is very tall, just start on new page and let it overflow

        self.set_fill_color(*CODE_BG)
        # Draw background (may need multiple pages for very tall blocks)
        draw_h = min(block_h, self.h - y_start - 10)
        self.rect(self.l_margin, y_start, self.w - self.l_margin - self.r_margin, draw_h, "F")
        self.set_xy(self.l_margin + 4, y_start + 4)
        self.set_font(F_CODE, "", 7.5)
        self.set_text_color(*CODE_FG)
        for line in lines:
            if len(line) > 105:
                line = line[:102] + "..."
            if self.get_y() > self.h - 20:
                self.add_page()
                y_start = self.get_y()
                remaining = len(lines) - lines.index(line)
                rem_h = remaining * line_h + 6
                self.set_fill_color(*CODE_BG)
                self.rect(self.l_margin, y_start, self.w - self.l_margin - self.r_margin, min(rem_h, self.h - y_start - 10), "F")
                self.set_xy(self.l_margin + 4, y_start + 2)
                self.set_font(F_CODE, "", 7.5)
                self.set_text_color(*CODE_FG)
            self.cell(0, line_h, line, new_x="LMARGIN", new_y="NEXT")
            self.set_x(self.l_margin + 4)
        self.ln(6)

    def write_bullet(self, text, indent=0):
        x_offset = self.l_margin + 4 + indent * 8
        self.set_x(x_offset)
        self.set_font(F_BODY, "", 10)
        self.set_text_color(*DARK)
        self.write(5.5, "\u2022 ")
        self._write_rich_text(text)
        self.ln(1)

    def write_numbered(self, num, text):
        self.set_x(self.l_margin + 4)
        self.set_font(F_BODY, "B", 10)
        self.set_text_color(*BLUE)
        self.write(5.5, f"{num}. ")
        self.set_font(F_BODY, "", 10)
        self.set_text_color(*DARK)
        self._write_rich_text(text)
        self.ln(1)

    def write_table(self, headers, rows):
        self.ln(2)
        col_count = len(headers)
        available_w = self.w - self.l_margin - self.r_margin

        # Calculate column widths based on content
        col_widths = []
        for i in range(col_count):
            max_len = len(headers[i])
            for row in rows:
                if i < len(row):
                    max_len = max(max_len, len(row[i]))
            col_widths.append(max(max_len, 3))
        total = sum(col_widths)
        col_widths = [w / total * available_w for w in col_widths]
        col_widths = [max(w, 15) for w in col_widths]
        total = sum(col_widths)
        col_widths = [w / total * available_w for w in col_widths]

        # Check if table fits on page
        est_h = (len(rows) + 1) * 7
        if self.get_y() + est_h > self.h - 20 and est_h < self.h - 40:
            self.add_page()

        # Header row
        self.set_font(F_BODY, "B", 8.5)
        self.set_fill_color(*TABLE_HEADER_BG)
        self.set_text_color(*WHITE)
        for i, h in enumerate(headers):
            self.cell(col_widths[i], 7, h.strip(), border=0, fill=True)
        self.ln()

        # Data rows
        self.set_font(F_BODY, "", 8.5)
        self.set_text_color(*DARK)
        for ri, row in enumerate(rows):
            if ri % 2 == 1:
                self.set_fill_color(*TABLE_ALT_BG)
            else:
                self.set_fill_color(*WHITE)
            for i in range(col_count):
                val = row[i].strip() if i < len(row) else ""
                self.cell(col_widths[i], 6.5, val, border=0, fill=True)
            self.ln()
        self.ln(3)

    def write_hr(self):
        self.ln(4)
        self.set_draw_color(220, 220, 220)
        self.set_line_width(0.5)
        y = self.get_y()
        self.line(self.l_margin, y, self.w - self.r_margin, y)
        self.ln(6)


def parse_table(lines):
    """Parse markdown table lines into headers and rows."""
    headers = [c.strip() for c in lines[0].strip().strip("|").split("|")]
    rows = []
    for line in lines[2:]:  # skip separator line
        cells = [c.strip() for c in line.strip().strip("|").split("|")]
        rows.append(cells)
    return headers, rows


def build_pdf():
    with open(MD_PATH, "r") as f:
        raw_lines = f.readlines()

    pdf = ProjectPDF()
    pdf.alias_nb_pages()
    pdf.add_page()

    i = 0
    is_first_h1 = True
    in_code = False
    code_lines = []

    while i < len(raw_lines):
        line = raw_lines[i].rstrip("\n")

        # Code block start/end
        if line.startswith("```"):
            if in_code:
                pdf.write_code_block(code_lines)
                code_lines = []
                in_code = False
            else:
                in_code = True
            i += 1
            continue

        if in_code:
            code_lines.append(line)
            i += 1
            continue

        # Blank line
        if line.strip() == "":
            i += 1
            continue

        # HR
        if line.strip() == "---":
            pdf.write_hr()
            i += 1
            continue

        # H1
        if line.startswith("# "):
            text = line[2:].strip()
            if is_first_h1:
                pdf.write_title(text)
                is_first_h1 = False
            else:
                pdf.write_title(text)
            i += 1
            continue

        # H2
        if line.startswith("## "):
            text = line[3:].strip()
            pdf.write_h2(text)
            i += 1
            continue

        # H3
        if line.startswith("### "):
            text = line[4:].strip()
            pdf.write_h3(text)
            i += 1
            continue

        # Table detection
        if "|" in line and i + 1 < len(raw_lines) and re.match(r'\s*\|[\s\-:|]+\|', raw_lines[i + 1]):
            table_lines = []
            while i < len(raw_lines) and "|" in raw_lines[i].strip():
                table_lines.append(raw_lines[i].rstrip("\n"))
                i += 1
            headers, rows = parse_table(table_lines)
            pdf.write_table(headers, rows)
            continue

        # Blockquote
        if line.startswith("> "):
            text = line[2:].strip()
            while i + 1 < len(raw_lines) and raw_lines[i + 1].startswith("> "):
                i += 1
                text += " " + raw_lines[i].rstrip("\n")[2:].strip()
            pdf.write_blockquote(text)
            i += 1
            continue

        # Bullet
        if line.startswith("- ") or line.startswith("* "):
            text = line[2:].strip()
            pdf.write_bullet(text)
            i += 1
            continue

        # Numbered list
        m = re.match(r'^(\d+)\.\s+(.*)', line)
        if m:
            num = m.group(1)
            text = m.group(2).strip()
            pdf.write_numbered(num, text)
            i += 1
            continue

        # Regular paragraph
        para = line.strip()
        while i + 1 < len(raw_lines):
            next_line = raw_lines[i + 1].rstrip("\n")
            if (next_line.strip() == "" or next_line.startswith("#") or
                    next_line.startswith("```") or next_line.startswith("- ") or
                    next_line.startswith("* ") or next_line.startswith("> ") or
                    next_line.strip() == "---" or
                    re.match(r'^\d+\.\s+', next_line) or
                    ("|" in next_line and i + 2 < len(raw_lines) and
                     re.match(r'\s*\|[\s\-:|]+\|', raw_lines[i + 2]))):
                break
            i += 1
            para += " " + next_line.strip()
        pdf.write_paragraph(para)
        i += 1

    pdf.output(PDF_PATH)
    print(f"PDF generated: {PDF_PATH}")


if __name__ == "__main__":
    build_pdf()
