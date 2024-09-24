# You are Alice and want to send 2000 kr. to Bob through a confidential
# message. You decide to use the ElGamal public key method.
# The keying material you should use to send the message to Bob is as
# follows:
#     • The shared base g=666
#     • The shared prime p=6661
#     • Bob’s public key P K = gx mod p =2227
# Send the message ’2000’ to Bob.

# 2. You are now Eve and intercept Alice’s encrypted message. Find Bob’s
# private key and reconstruct Alice’s message.

# 3. You are now Mallory and intercept Alice’s encrypted message. However,
# you run on a constrained device and are unable to find Bob’s private key.
# Modify Alice’s encrypted message so that when Bob decrypts it, he will
# get the message ’6000’.

import random

# Given values
g = 666 # shared base
p = 6661 # shared prime
pk = 2227 # Bob's public key
message = 2000 # Alice's message to Bob

# test example
# g = 60 # shared base
# p = 283 # shared prime
# pk = 216 # Bob's public key
# message = 101 # Alice's message to Bob


def sendMessage(): 
    print("1. Send message")
    # test k example
    # k = 36

    # Step 1: Choose a random integer k such that 1 <= k <= p-2
    # note: each time a message is sent, k is different each time (or at least random)
    k = random.randint(1, p-2)


    # Step 2: Compute c1 = g^k mod p
    # c1= pow(g, k, p) can also be used
    # c1 = 666^k mod 6661
    c1 = pow(g, k) % p

    # Step 3: Compute c2 = message * pk^g mod p
    # c2 = pow(pk, k, p) can also be used
    c2 = (message * pow(pk, k)) % p

    # Step 4: "Return" the ciphertext which is the pair of c1, c2
    ciphertext = (c1, c2)
    print(f"Alice: {ciphertext}")
    
    return ciphertext

    
    

def interceptMessage():
    print("2. Get Alice's message")
    privatekey = -1 # Bob's privatekey 
    
    # Step 1. Recreate Bob's public key to find the private key
    # privateKey = g^i mod p
    # privateKey = 666^[0-6661] mod 6661 
    # This is a brute force way to get the private key
    # If you know the shared base and the shared prime, you can then slowly find the private key
    for i in range(p):
        # this is the same way a public key is created
        checkKey = pow(g, i) % p
        if checkKey == pk: 
            privatekey = i
            
    print(f"Bob's private key is: {privatekey}")
    
    # Step 2. return the found private key, and then decrypt the message
    return privatekey

def modifyMessage(message):
    print("3. Modify Alice's message")
    c1, c2 = message
    
    #c2 is the message wearing a mask, so modifing the mask will change the message
    # this is called a malleability attack. normaly you wouldnt know what is in the message before hand.
    # But even with knowning the message, it is still the same aprouch.
    return (c1, (c2*3))
    

def decrypt(message, key):
    c1, c2 = message   
    
    computedC1 = pow(c1, (p - key - 1))
    
    msg = (computedC1 * c2) % p 
    
    return msg
    
    
    
    
    
# Send the message ’2000’ to Bob.
cipherMessage = sendMessage()
print("")

# Find Bob’s private key and reconstruct Alice’s message.
bobprivatekey = interceptMessage()
decryptMsg = decrypt(cipherMessage, bobprivatekey)
print(f"Eve decrypted the message to be: {decryptMsg}")
print("")

# Modify Alice’s encrypted message, so when decrypted, it gives 6000
modifiedMessage = modifyMessage(cipherMessage)
decryptMsg = decrypt(modifiedMessage, bobprivatekey)
print(f"Alice: {decryptMsg}")