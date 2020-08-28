import re
import logging
import os

kv_pat = re.compile(r"^\s*(?P<attr>\w+)(-(?P<class>[\w,]+))?\s*=\s*(?P<value>.*)$")
ignore_pat = re.compile(r"^\s*#+|^\s*$")
stanza_pat = re.compile(r"^\s*\[\s*((?P<type>\w+)::)?(?P<pat>[\w\-.()?+-\\/*{}|!~#$]+)\s*\]")



class Output(object):

    def __init__(self, name, pat, stanta_type):
        self._setting = dict()
        self.name = name
        self.pat = pat
        self.type = stanta_type


    # def __setattr__(self, attr, value):
    #     self._setting[attr] = value

    def __getattr__(self, attr):
        return self._setting.get(attr)

    def set_attr(self, attr, value):
        self._setting[attr] = value

    def has_attr(self, attr):
        if self._setting.get(attr) is not None:
            return True
        return False

    def setting_keys(self):
        return self._setting.keys()

    
    @staticmethod
    def parse(file_path):
        with open(file_path) as fin:
            text = fin.readlines()
        
        result = dict()
        cur = Output(name="default", pat=None, stanta_type='default')

        for l in text:
            if re.match(ignore_pat, l):
                continue

            m = re.match(stanza_pat, l)
            if not m:
                m = re.match(kv_pat, l)
                if not m:
                    logging.warning("Can't distinguish line: %s", l)
                    continue

                attr = m.groupdict()['attr']
                value = m.groupdict()['value']
                cur.set_attr(attr, value)


            else:
                if result.get(cur.name) is not None:
                    # union
                    for k,v in cur._setting:
                        result[cur.name].set_attr(k,v)
                else:
                    result[cur.name] = cur

                m = m.groupdict()
                if m['type'] != None:
                    name = "{0}::{1}".format(m['type'], m['pat'])
                else:
                    name = m['pat']
                stanza_type = m['type'] if not m['type'] else "sourcetype"
                pat = m['pat']
                cur = Output(name, stanza_type, pat)


        return result

    @staticmethod
    def graph(outputs):
        result = "subgraph clusteroutputs{\nlabel=\"outputs\";bgcolor=\"mintcream\";node [fontname = \"Verdana\", fontsize = 10, color=\"green\", shape=\"record\"];\n"
        for output in outputs:
            result += "output_" + outputs[output].name + "\n"
        result += "};\n"
        return result




class Transform(object):
    def __init__(self, name, pat, stanta_type):
        self._setting = dict()
        self.name = name
        self.pat = pat
        self.type = stanta_type


    # def __setattr__(self, attr, value):
    #     self._setting[attr] = value

    def __getattr__(self, attr):
        return self._setting.get(attr)

    def set_attr(self, attr, value):
        self._setting[attr] = value

    def has_attr(self, attr):
        if self._setting.get(attr) is not None:
            return True
        return False

    def setting_keys(self):
        return self._setting.keys()

    def graph_self(self):
        fields = ""
        # regex = re.escape(self.REGEX)

        if self.FORMAT is not None:
            for field in self.FORMAT:
                if field[1] == "":
                    fields = field[0]
                    break

                fields += "=".join(field)
                fields += "|"
        else:
            try:
                groups = re.compile(self.REGEX).groupindex
            except Exception as e:
                # logging.error("regex compile error, name: %s, regex: %s", self.name, self.REGEX)
                # return "{0} [label=\"{{{0}|REGEX= {1}|REGEX_ERROR}}\", color=\"red\"];\n".format(self.name,regex)
                return "transforms_{0} [label=\"{{transforms_{0}|REGEX_ERROR}}\", color=\"red\"];\n".format(self.name)
            for field in groups:
                fields += "{0}|".format(field)

        fields = fields.rstrip("|")
        if fields != "":
            # node = "{0} [label=\"{{{0}|REGEX={1}|{2}}}\"];\n".format(self.name, regex, fields)
            node = "transforms_{0} [label=\"{{transforms_{0}|{1}}}\"];\n".format(self.name, fields)
        else:
            # node = "{0} [label=\"{{{0}|REGEX={1}}}\"];\n".format(self.name, regex)
            node = "transforms_{0} [label=\"{{transforms_{0}}}\"];\n".format(self.name)


        return node


    @staticmethod
    def graph(transforms):
        result = "subgraph clustertransforms{\nlabel=\"transforms\";bgcolor=\"mintcream\";node [fontname = \"Verdana\", fontsize = 10, color=\"skyblue\", shape=\"record\"];\n"
        for transform in transforms:
            result += transforms[transform].graph_self()
        result += "};\n"
        return result


    @staticmethod
    def parse(file_path):
        with open(file_path) as fin:
            text = fin.readlines()
        
        result = dict()
        cur = Transform(name="default", pat=None, stanta_type='default')
        line_no = 0

        for l in text:
            line_no +=1
            if re.match(ignore_pat, l):
                continue

            m = re.match(stanza_pat, l)
            if not m:
                m = re.match(kv_pat, l)
                if not m:
                    logging.warning("Can't distinguish line: %s", l)
                    continue

                attr = m.groupdict()['attr']
                value = m.groupdict()['value']

               
                if attr =='FORMAT':
                    field_pat = re.compile(r"\s*(?P<type>\w+)::(?P<pat>\$\w+)\s*")
                    fields = re.findall(field_pat, value)
                    if not fields:
                        if value != "":
                            fields=((value, ""),)
                        else:
                            logging.warning("Can't distinguish FORMAT in line %d: %s",line_no, l)
                            continue

                    regex = list()
                    for i in fields:
                        regex.append(i)
                    value = regex

                if attr == "REGEX":
                    value = value.replace("?<", "?P<")

                cur.set_attr(attr, value)


            else:
                if result.get(cur.name) is not None:
                    # union
                    for k in cur._setting:
                        v = cur.get(k)
                        result[cur.name].set_attr(k,v)
                else:
                    result[cur.name] = cur

                m = m.groupdict()
                if m['type'] != None:
                    name = "{0}::{1}".format(m['type'], m['pat'])
                else:
                    name = m['pat']
                stanza_type = m['type'] if not m['type'] else "sourcetype"
                pat = m['pat']
                cur = Transform(name, stanza_type, pat)

        # write final stanza
        if result.get(cur.name) is not None:
            # union
            for k in cur._setting:
                v = cur.get(k)
                result[cur.name].set_attr(k,v)
        else:
            result[cur.name] = cur


        return result


class Prop(object):
    """setting in props.conf"""
    def __init__(self, name,stanta_type, pat):
        self._setting = dict()
        self.name = name
        self.pat = pat
        self.type = stanta_type



    def __getattr__(self, attr):
        return self._setting.get(attr)

    def get(self, attr):
        return self._setting.get(attr)

    def set_attr(self, attr, value):
        self._setting[attr] = value

    def has_attr(self, attr):
        if self._setting.get(attr) is not None:
            return True
        return False

    def setting_keys():
        return self._setting.keys()

    def graph_self(self):
        connect = ""
        props_graph = ""
        transform_label = ""
        report_label = ""
        tran_classes = self.TRANSFORMS
        rep_classes = self.REPORT

        name = self.name.replace("-","_")

        if tran_classes:
            for c in tran_classes:
                transforms = tran_classes[c]
                
                for t in transforms:
                    transform_label += "<transforms_{0}>transforms_{0}".format(t)
                    connect += "\"props_{0}\" -> transforms_{1};\n".format(name, t.replace("-","_"))
                    # connect += "\"props_{0}\":{1} -> transforms_{1};\n".format(self.name, t)
                transform_label = transform_label.strip("|")

        if rep_classes:
            for c in rep_classes:
                reports = rep_classes[c]
                
                for t in reports:
                    report_label += "<transforms_{0}>transforms_{0}".format(t)
                    connect += "\"props_{0}\" -> transforms_{1} [stype=\"dashed\",color=\"green\" label=\"REPORT\"];\n".format(name, t.replace("-","_"))
                    # connect += "\"props_{0}\":{1} -> transforms_{1};\n".format(self.name, t)
                transform_label = transform_label.strip("|")

        # props_graph = "\"props_{0}\" [label=\"{{{0}|{1}}}\"];\n".format(self.name, transform_label)
        # props_graph = "\"props_{0}\" [label=\"{{{0}}}\"];\n".format(self.name)
        props_graph = "\"props_{0}\";\n".format(name)

        return props_graph, connect




    @staticmethod
    def graph(props):
        result = "subgraph clusterpropos{\nlabel=\"propos\";bgcolor=\"mintcream\";node [fontname = \"Verdana\", fontsize = 10, color=\"yellow\", shape=\"record\"];\n"
        connects = ""
        for prop in props:
            props_graph, connect = props[prop].graph_self()
            result += props_graph
            connects += connect
        result += "};\n"
        return result, connects


    @staticmethod
    def parse(file_path):
        with open(file_path) as fin:
            text = fin.readlines()
        
        result = dict()
        cur = Prop(name="default", pat=None, stanta_type='default')

        for l in text:
            if re.match(ignore_pat, l):
                continue

            m = re.match(stanza_pat, l)
            if not m:
                m = re.match(kv_pat, l)
                if not m:
                    logging.warning("Can't distinguish kv: %s", l)
                    continue

                attr = m.groupdict()['attr']
                value = m.groupdict()['value']
                ns_class = m.groupdict()['class']
                if ns_class is None:
                    ns_class = "_DEFAULT"

                if attr == "TRANSFORMS" or attr == "REPORT":
                    value = value.split(",")
                    transforms = list(map(lambda x: x.strip(), value))
                    if not transforms :
                        logging.error("Can't distinguish TRANSFORM: %s", l)
                        continue

                    old = cur.get(attr)
                    if old is not None and old.get(ns_clas) is not None:
                        old[ns_class].append(transforms)
                        continue
                    elif old is not None :
                        old[ns_class] = transforms
                        continue

                    value = {ns_class: transforms}
                elif attr == "EXTRACT":
                    value = value.split(",")
                    transforms = list(map(lambda x: x.strip(), value))
                    if not transforms :
                        logging.error("Can't distinguish EXTRACT: %s", l)
                        continue

                    old = cur.get(attr)
                    if old is not None and old.get(ns_class) is not None:
                        old[ns_class].append(transforms)
                        continue
                    elif old is not None :
                        old[ns_class] = transforms
                        continue

                    value = {ns_class: transforms}


                cur.set_attr(attr, value)


            else:
                ## new stanza,write old stanza
                if result.get(cur.name) is not None:
                    # union
                    old = result[cur.name]
                    if old.priority is None or old.priority <= cur.priority:
                        for k in cur._setting:
                            v = cur.get(k)
                            result[cur.name].set_attr(k,v)
                    else:
                        for k in old._setting:
                            v = old.get(k)
                            cur.set_attr(k,v)
                        result[cur.name] = cur

                else:
                    result[cur.name] = cur
                    
                m = m.groupdict()
                if m['type'] != None:
                    name = "{0}::{1}".format(m['type'], m['pat'])
                else:
                    name = m['pat']

                stanza_type = m['type'] if m['type'] is not None else "sourcetype"
                pat = m['pat']
                cur = Prop(name, stanza_type, pat)

        ## write final stanza
        if result.get(cur.name) is not None:
            # union
            old = result[cur.name]
            if old.priority is None or old.priority <= cur.priority:
                for k in cur._setting:
                    v = cur.get(k)
                    result[cur.name].set_attr(k,v)
            else:
                for k in old._setting:
                    v = old.get(k)
                    cur.set_attr(k,v)
                result[cur.name] = cur

        else:
            result[cur.name] = cur


        return result



class Input(object):
    """setting in props.conf"""
    def __init__(self, name, stanta_type, pat):
        self._setting = dict()
        self.name = name
        self.pat = pat
        self.type = stanta_type




    def __getattr__(self, attr):
        return self._setting.get(attr)

    def get(self, attr):
        return self._setting.get(attr)

    def set_attr(self, attr, value):
        self._setting[attr] = value

    def has_attr(self, attr):
        if self._setting.get(attr) is not None:
            return True
        return False

    def setting_keys():
        return self._setting.keys()

    def graph_self(self, props):
        connect = ""
        name = self.name.replace("-","_")
        input_graph = "\"inputs_{0}\";\n".format(name)

        for k in props:
            prop = props[k]
            if self.host is not None and prop.type == "host" and prop.pat == self.host:
                connect += "\"inputs_{0}\" -> props_{1} [stype=\"dashed\",color=\"blue\" label=\"host\"];\n".format(name, prop.name.replace("-","_"))
            elif self.source is not None and prop.type == "source" and prop.pat == self.source:
                connect += "\"inputs_{0}\" -> props_{1} [stype=\"dashed\",color=\"blue\" label=\"source\"];\n".format(name, prop.name.replace("-","_"))
            elif self.sourcetype is not None and prop.type == "sourcetype" and prop.pat == self.sourcetype:
                connect += "\"inputs_{0}\" -> props_{1} [stype=\"dashed\",color=\"blue\" label=\"sourcetype\"];\n".format(name, prop.name.replace("-","_"))

        return input_graph,connect

    @staticmethod
    def graph(inputs, props):
        result = "subgraph clusterinputs{\nlabel=\"inputs\";bgcolor=\"mintcream\";node [fontname = \"Verdana\", fontsize = 10, color=\"black\", shape=\"record\"];\n"
        connects = ""
        for input in inputs:
            props_graph, connect = inputs[input].graph_self(props)
            result += props_graph
            connects += connect
        result += "};\n"
        return result, connects


    @staticmethod
    def parse(file_path):
        with open(file_path) as fin:
            text = fin.readlines()
        
        result = dict()
        cur = Input(name="default", pat=None, stanta_type='default')
        input_stanza_pat = re.compile(r"\s*\[\s*((?P<type>\w+)::)?(?P<pat>[\w\-.()?+-\\/*{}|!~#$]+)\s*\]")

        for l in text:
            if re.match(ignore_pat, l):
                continue

            m = re.match(input_stanza_pat, l)
            if not m:
                m = re.match(kv_pat, l)
                if not m:
                    logging.warning("Can't distinguish kv: %s", l)
                    continue

                attr = m.groupdict()['attr']
                value = m.groupdict()['value']

                cur.set_attr(attr, value)


            else:
                ## new stanza,write cur stanza
                if result.get(cur.name) is not None:
                    # union
                    old = result[cur.name]
                    if old.priority is None or old.priority <= cur.priority:
                        for k in cur._setting:
                            v = cur.get(k)
                            result[cur.name].set_attr(k,v)
                    else:
                        for k in old._setting:
                            v = old.get(k)
                            cur.set_attr(k,v)
                        result[cur.name] = cur

                else:
                    result[cur.name] = cur
                    
                # cur point to new stanza
                m = m.groupdict()
                if m.get('type') is not None:
                    name = "{0}::{1}".format(m['type'], m['pat'])
                else:
                    name = m['pat']

                stanza_type = m['type'] if not m['type'] else "unknow"
                pat = m['pat']
                cur = Input(name, stanza_type, pat)

        ## write final stanza
        if result.get(cur.name) is not None:
            # union
            old = result[cur.name]
            if old.priority <= cur.priority:
                for k,v in cur._setting:
                    result[cur.name].set_attr(k,v)
            else:
                for k,v in old._setting:
                    cur.set_attr(k,v)
                result[cur.name] = cur

        else:
            result[cur.name] = cur

        return result



if __name__ == "__main__":
    # app_path = "E:\\splunk-conf\\etc\\system\\default"
    app_path = "E:\\todo\\spl_examples\\default"
    # outputs = Output.parse(os.path.join(app_path,"outputs.conf"))
    transforms = Transform.parse(os.path.join(app_path,"transforms.conf"))
    props = Prop.parse(os.path.join(app_path,"props.conf"))
    # inputs = Input.parse(os.path.join(app_path,"inputs.conf"))

    with open("E:\\first.dot", "w") as fout:
        # outputs_graph = Output.graph(outputs).replace("-","_")
        transforms_graph = Transform.graph(transforms).replace("-","_")
        props_graph, connect = Prop.graph(props)
        props_graph = props_graph.replace("-","_")
        # input_graph, input_connect = Input.graph(inputs, props)

        fout.write("digraph data_flow{\nrankdir=\"LR\";\nsplines = \"polyline\";\n")
        # fout.write(outputs_graph)
        fout.write(transforms_graph)
        fout.write(props_graph)
        # fout.write(input_graph)
        fout.write(connect)
        # fout.write(input_connect)
        fout.write("}")
