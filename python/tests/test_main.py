import unittest
from unittest.mock import patch

import main


class FakeResponse:
    """Minimal stand-in for the parts of requests.Response used by main."""

    def __init__(self, payload=None, url=None):
        self.payload = payload
        self.url = url

    def raise_for_status(self):
        pass

    def json(self):
        return self.payload


class StopAfterPolls:
    """Stop the monitor after a fixed number of polling cycles."""

    def __init__(self, polls):
        self.polls = polls
        self.completed_polls = 0

    def is_set(self):
        return self.completed_polls >= self.polls

    def wait(self, timeout):
        self.completed_polls += 1
        return self.is_set()


class MainTests(unittest.TestCase):
    def test_get_m2o_stream_resolves_the_redirected_audio_url(self):
        metadata = FakeResponse({"audio_url": "https://audio.example/playlist"})
        stream = FakeResponse(url="https://cdn.example/live")

        with patch("main.requests.get", side_effect=[metadata, stream]) as get:
            result = main.get_m2o_stream()

        self.assertEqual(result, "https://cdn.example/live")
        self.assertEqual(get.call_count, 2)

    def test_get_track_formats_artist_and_song(self):
        response = FakeResponse({"title": "Take On Me A-HA"})

        with patch("main.requests.get", return_value=response):
            result = main.get_track("https://api.example/onair")

        self.assertEqual(result, "A-Ha - Take On Me")

    def test_monitor_prints_only_new_tracks_with_a_timestamp(self):
        stop_event = StopAfterPolls(polls=3)

        with patch(
            "main.get_track",
            side_effect=["Artist - First", "Artist - First", "Artist - Second"],
        ), patch("main.datetime") as datetime_mock, patch("builtins.print") as print_mock:
            datetime_mock.now.return_value.strftime.return_value = "09:45 16.10.2026"

            main.monitor_track(stop_event)

        self.assertEqual(
            print_mock.call_args_list,
            [
                unittest.mock.call("09:45 16.10.2026 Artist - First"),
                unittest.mock.call("09:45 16.10.2026 Artist - Second"),
            ],
        )


if __name__ == "__main__":
    unittest.main()
