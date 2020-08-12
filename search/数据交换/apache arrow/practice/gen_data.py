# -*- coding: utf-8 -*-
"""
    File Name: gen_data
    Description: ""
    Author: haha.gan
    Date: 2020/6/28 16:43
"""

import json
import os


def load_data(f_name):
    with open(f_name, "r") as fin:
        return json.load(fin)


def expand_data(data, times, f_name):
    with open(os.path.join("json_data", f_name), "w") as fout:
        data['hits']['hits'] *= times
        json.dump(data, fout)


if __name__ == '__main__':
    data = load_data("test1.json")
    expand_data(data, 1000, "test1.json")