import os
import pytest
from extract import extract_text, ExtractionError

FIXTURES = os.path.join(os.path.dirname(__file__), "test_fixtures")


def _joined(pages):
    return "\n".join(text for _, text in pages)


def test_extract_plaintext(tmp_path):
    p = tmp_path / "doc.txt"
    p.write_text("Hello world.\nThis is a plain document.")
    pages = extract_text(str(p), "text/plain")
    assert pages == [(1, "Hello world.\nThis is a plain document.")]


def test_extract_markdown(tmp_path):
    p = tmp_path / "doc.md"
    p.write_text("# Title\n\nBody paragraph.")
    pages = extract_text(str(p), "text/markdown")
    assert len(pages) == 1
    assert pages[0][0] == 1
    assert "Title" in pages[0][1] and "Body paragraph" in pages[0][1]


def test_extract_unknown_returns_empty(tmp_path):
    p = tmp_path / "doc.bin"
    p.write_bytes(b"\x00\x01\x02\x03")
    with pytest.raises(ExtractionError):
        extract_text(str(p), "application/octet-stream")


def _has_fixture(name: str) -> bool:
    return os.path.exists(os.path.join(FIXTURES, name))


@pytest.mark.skipif(not _has_fixture("sample.pdf"), reason="fixture missing; run test_fixtures/_generate.py")
def test_extract_pdf():
    pages = extract_text(os.path.join(FIXTURES, "sample.pdf"), "application/pdf")
    assert "semantic search test marker" in _joined(pages)
    # PDF pages should be 1-indexed and contiguous starting at 1.
    page_nums = [p for p, _ in pages]
    assert page_nums == list(range(1, len(page_nums) + 1))


@pytest.mark.skipif(not _has_fixture("encrypted.pdf"), reason="fixture missing; run test_fixtures/_generate.py")
def test_extract_pdf_encrypted_raises():
    with pytest.raises(ExtractionError):
        extract_text(os.path.join(FIXTURES, "encrypted.pdf"), "application/pdf")


@pytest.mark.skipif(not _has_fixture("sample.docx"), reason="fixture missing; run test_fixtures/_generate.py")
def test_extract_docx():
    pages = extract_text(os.path.join(FIXTURES, "sample.docx"))
    assert pages and pages[0][0] == 1
    assert "document marker phrase" in _joined(pages)


@pytest.mark.skipif(not _has_fixture("sample.xlsx"), reason="fixture missing; run test_fixtures/_generate.py")
def test_extract_xlsx():
    pages = extract_text(os.path.join(FIXTURES, "sample.xlsx"))
    assert "spreadsheet marker" in _joined(pages)
    # Each sheet emits its own page.
    assert all(p >= 1 for p, _ in pages)


@pytest.mark.skipif(not _has_fixture("sample.pptx"), reason="fixture missing; run test_fixtures/_generate.py")
def test_extract_pptx():
    pages = extract_text(os.path.join(FIXTURES, "sample.pptx"))
    assert "slide marker" in _joined(pages)
    assert all(p >= 1 for p, _ in pages)


@pytest.mark.skipif(not _has_fixture("sample.png"), reason="fixture missing; run test_fixtures/_generate.py")
def test_extract_image_ocr():
    try:
        pages = extract_text(os.path.join(FIXTURES, "sample.png"), "image/png")
    except ExtractionError as e:
        if "OCR unavailable" in str(e):
            pytest.skip(f"tesseract/pytesseract not installed: {e}")
        raise
    assert pages and pages[0][0] == 1
    assert "image OCR marker" in _joined(pages).replace("\n", " ")


def test_extract_and_embed_endpoint(tmp_path):
    if not os.environ.get("EMBED_TEST_LIVE"):
        pytest.skip("set EMBED_TEST_LIVE=1 to run live model tests")

    from main import app
    client = app.test_client()

    src = tmp_path / "doc.txt"
    src.write_text("The quick brown fox jumps over the lazy dog. " * 30)

    resp = client.post("/extract-and-embed", json={
        "path": str(src),
        "mimeHint": "text/plain",
    })
    assert resp.status_code == 200
    body = resp.get_json()
    assert isinstance(body, list)
    assert len(body) >= 1
    assert all({"vector", "chunkIndex", "snippet", "page"} <= c.keys() for c in body)
    assert all(len(c["vector"]) == 1024 for c in body)
    assert all(c["page"] >= 1 for c in body)
