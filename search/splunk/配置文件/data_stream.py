import re


class Output(object):

    def __init__(self, name=None):
        self.name = name
    
    @staticmethod
    def parse(file_path):
        with open(file_path) as fin:
            text = fin.readlines()
        
        result = list()
        cur = Output()
        groupname_pat = re.compile(r"\s*\[\s*tcpout:\s*(?P<name>[a-z]+)\s*\]|\s*defaultGroup\s*=\s*(?P<names>\S+)")

        for l in text:
            m = re.match(groupname_pat, l)
            if not m:
                continue
            else:
                m = m.groupdict()

            if m['name']:
                result.append(Output(m['name']))
            elif m['names']:
                names = m['names'].split(',')
                for name in names:
                    result.append(Output(name.strip()))
        return result


class Transform(object):
    def __init__(self, pat, type):
        self.pat = pat
        self.type = type


    @staticmethod
    def parse(file_path):
        with open(file_path) as fin:
            text = fin.readlines()
        
        result = list()
        cur = Transform(pat=None, type='defalut')
        groupname_pat = re.compile(r"\s*\[\s*((?P<type>\S+)::)?(?P<pat>\w+)\s*\]")

        for l in text:
            m = re.match(groupname_pat, l)
            if not m:
                continue
            else:
                m = m.groupdict()
                cur = Transform(pat)

        return result


if __name__ == "__main__":
    print(Output.parse("E:\\splunk-conf\\etc\\system\\default\\outputs.conf"))
