import json
import pytest
import main


@pytest.fixture
def client():
    main.app.config["TESTING"] = True
    return main.app.test_client()


def test_encode_text_returns_plain_and_image_vectors(client):
    resp = client.post("/encode-text", json={"text": "soc2"})
    assert resp.status_code == 200
    body = json.loads(resp.data)
    assert len(body["text_features"]) == main.EMBEDDING_DIM
    assert len(body["image_query_features"]) == main.EMBEDDING_DIM
    # The plain and prompted encodings must differ — that is the whole point.
    assert body["text_features"] != body["image_query_features"]
