# -*- coding: utf-8 -*-
"""
    File Name: pandas_read
    Description: ""
    Author: haha.gan
    Date: 2020/6/22 11:20
"""
import pandas
import pdb;

p = pandas.Series([1, 2, 3, 4])
p1 = pandas.Series([5, 6, 7, p])
# print(p[0])
print(p1[3])
d = pandas.DataFrame()
