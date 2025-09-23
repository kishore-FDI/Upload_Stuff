a=input()
bit = 1
ans = 0
for i in range(1,31):
    bit=1<<i
    if(bit>int(a)): break
    s = str(bit)
    idx = 0
    for j in a:
        if j==s[idx]:
            idx+=1
        if idx==len(s):
            ans = max(ans,int(s))
            break
print(ans)