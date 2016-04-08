def set_key(d, key, value):
    if d is None:
        d = dict()
    d[key] = value
    return d


class FilterModule(object):
    def filters(self):
        return {'set_key': set_key}
