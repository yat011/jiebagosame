import jieba

if __name__ == '__main__':
    print(jieba.__version__)
    result = []
    jieba.add_word('__解释性__')
    with open("./example.txt",encoding='utf-8') as f:
        for line in f:
            result.append(list(jieba.cut(line.strip(),HMM=False)))

    with open("./py_ans.txt","w", encoding="utf-8") as f:
        for line in result:
            f.write("|".join(line)+"\n")