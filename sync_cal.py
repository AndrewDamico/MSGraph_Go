import ctypes
import os
import pathlib


def sync(dir):
    path = os.path.join(dir,"sync.so")
    cur = pathlib.Path().resolve()
    lib = ctypes.CDLL(path)
    go = lib.main

    try:
        go()
    except:
        print("error executing Go command")
        pass


if __name__ == "__main__":
    dir = pathlib.Path().resolve()
    sync(dir)
