import math

data = []
with open("map2.txt", "r") as f:
    lines = f.readlines()
    for line in lines:
        tmp = line.split(",")
        y, x = float(tmp[0]), float(tmp[1])
        data.append((x, y))
with open("location2.sql", "w") as f:
    for item in data:
        x, y = item[0], item[1]
        f.write("insert into locations set x={},y={};\n".format(x,y))
