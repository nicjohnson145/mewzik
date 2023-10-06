from mutagen.easyid3 import EasyID3
import argparse as ap
from pydub import AudioSegment


def parse_args():
    parser = ap.ArgumentParser(description="Create a file with ID3 tags and random content")

    parser.add_argument("--artist", required=True)
    parser.add_argument("--album", required=True)
    parser.add_argument("--title", required=True)
    parser.add_argument("--track", type=int, required=True)

    parser.add_argument("--path", default="generated.mp3")

    return parser.parse_args()


def main():
    args = parse_args()

    # "touch" the file
    fl = AudioSegment.empty()
    fl.export(args.path, format="mp3")

    meta = EasyID3(args.path)
    meta["title"] = args.title
    meta["artist"] = args.artist
    meta["album"] = args.album
    meta["tracknumber"] = [args.track]
    meta.save(args.path, v1=2)


if __name__ == "__main__":
    main()
