import ctypes
import os
import pathlib


def sync(dir):
    print("loaded from:",__name__)
    print(dir)
    path = os.path.join(dir,"sync.so")
    print(path)
    cur = pathlib.Path().resolve()
    print ("Current Path:",cur)
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
