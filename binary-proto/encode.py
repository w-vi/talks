#! /usr/bin/env python3

import gzip
import json
import argparse
import msgpack
import cbor2


def main():
    usage = 'Usage: %prog [options]'
    parser = argparse.ArgumentParser()
    parser.add_argument(
        'filepath', help='Path to json file')

    args = parser.parse_args()

    with open(args.filepath) as json_file:
        data = json.load(json_file)

    msg = msgpack.packb(data, use_bin_type=True)

    with open("msgpack.dat", "wb") as msgpack_file:
        msgpack_file.write(msg)

    with gzip.GzipFile("json.gz", 'w') as fout:
        fout.write(json.dumps(data).encode('utf-8'))

    with open("cbordata.cbor", "wb") as cbor_file:
        cbor2.dump(data, cbor_file)

if __name__ == '__main__':
    main()
