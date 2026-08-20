package stanza_test

import (
	"gosrc.io/xmpp/stanza"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestPopEmptyQueue(t *testing.T) {
	var uaq stanza.UnAckQueue
	popped := uaq.Pop()
	if popped != nil {
		t.Fatalf("queue is empty but something was popped !")
	}
}

func TestPushUnack(t *testing.T) {
	uaq := initUnAckQueue()
	toPush := stanza.UnAckedStz{
		Id: 3,
		Stz: `<iq type='submit'
    from='confucius@scholars.lit/home'
    to='registrar.scholars.lit'
    id='kj3b157n'
    xml:lang='en'>
  <query xmlns='jabber:iq:register'>
    <username>confucius</username>
    <first>Qui</first>
    <last>Kong</last>
  </query>
</iq>`,
	}

	err := uaq.Push(&toPush)
	if err != nil {
		t.Fatalf("could not push element to the queue : %v", err)
	}

	if len(uaq.Uslice) != 4 {
		t.Fatalf("push to the non-acked queue failed")
	}
	for i := 0; i < 4; i++ {
		if uaq.Uslice[i].Id != i+1 {
			t.Fatalf("indexes were not updated correctly. Expected %d got %d", i, uaq.Uslice[i].Id)
		}
	}

	// Check that the queue is a fifo : popped element should not be the one we just pushed.
	popped := uaq.Pop()
	poppedElt, ok := popped.(*stanza.UnAckedStz)
	if !ok {
		t.Fatalf("popped element is not a *stanza.UnAckedStz")
	}

	if reflect.DeepEqual(*poppedElt, toPush) {
		t.Fatalf("pushed element is at the top of the fifo queue when it should be at the bottom")
	}

}

func TestPeekUnack(t *testing.T) {
	uaq := initUnAckQueue()

	expectedPeek := stanza.UnAckedStz{
		Id: 1,
		Stz: `<iq type='set'
    from='romeo@montague.net/home'
    to='characters.shakespeare.lit'
    id='search2'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'>
    <last>Capulet</last>
  </query>
</iq>`,
	}

	if !reflect.DeepEqual(expectedPeek, *uaq.Uslice[0]) {
		t.Fatalf("peek failed to return the correct stanza")
	}

}

func TestPeekNUnack(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)
	randPop := rand.Int31n(int32(initLen))

	peeked := uaq.PeekN(int(randPop))

	if len(uaq.Uslice) != initLen {
		t.Fatalf("queue length changed whith peek n operation : had %d found %d after peek", initLen, len(uaq.Uslice))
	}

	if len(peeked) != int(randPop) {
		t.Fatalf("did not peek the correct number of element from queue. Expected %d got %d", randPop, len(peeked))
	}
}

func TestPeekNUnackTooLong(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)

	// Have a random number of elements to peek that's greater than the queue size
	randPop := rand.Int31n(int32(initLen)) + 1 + int32(initLen)

	peeked := uaq.PeekN(int(randPop))

	if len(uaq.Uslice) != initLen {
		t.Fatalf("total length changed whith peek n operation : had %d found %d after pop", initLen, len(uaq.Uslice))
	}

	if len(peeked) != initLen {
		t.Fatalf("did not peek the correct number of element from queue. Expected %d got %d", initLen, len(peeked))
	}

}

func TestPopNUnack(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)
	randPop := rand.Int31n(int32(initLen))

	popped := uaq.PopN(int(randPop))

	if len(uaq.Uslice)+len(popped) != initLen {
		t.Fatalf("total length changed whith pop n operation : had %d found %d after pop", initLen, len(uaq.Uslice)+len(popped))
	}

	for _, elt := range popped {
		for _, oldElt := range uaq.Uslice {
			if reflect.DeepEqual(elt, oldElt) {
				t.Fatalf("pop n operation duplicated some elements")
			}
		}
	}
}

func TestPopNUnackTooLong(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)

	// Have a random number of elements to pop that's greater than the queue size
	randPop := rand.Int31n(int32(initLen)) + 1 + int32(initLen)

	popped := uaq.PopN(int(randPop))

	if len(uaq.Uslice)+len(popped) != initLen {
		t.Fatalf("total length changed whith pop n operation : had %d found %d after pop", initLen, len(uaq.Uslice)+len(popped))
	}

	for _, elt := range popped {
		for _, oldElt := range uaq.Uslice {
			if reflect.DeepEqual(elt, oldElt) {
				t.Fatalf("pop n operation duplicated some elements")
			}
		}
	}
}

func TestPopUnack(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)

	popped := uaq.Pop()

	if len(uaq.Uslice)+1 != initLen {
		t.Fatalf("total length changed whith pop operation : had %d found %d after pop", initLen, len(uaq.Uslice)+1)
	}
	for _, oldElt := range uaq.Uslice {
		if reflect.DeepEqual(popped, oldElt) {
			t.Fatalf("pop n operation duplicated some elements")
		}
	}

}

func initUnAckQueue() stanza.UnAckQueue {
	q := []*stanza.UnAckedStz{
		{
			Id: 1,
			Stz: `<iq type='set'
    from='romeo@montague.net/home'
    to='characters.shakespeare.lit'
    id='search2'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'>
    <last>Capulet</last>
  </query>
</iq>`,
		},
		{Id: 2,
			Stz: `<iq type='get'
    from='juliet@capulet.com/balcony'
    to='characters.shakespeare.lit'
    id='search3'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'/>
</iq>`},
		{Id: 3,
			Stz: `<iq type='set'
    from='juliet@capulet.com/balcony'
    to='characters.shakespeare.lit'
    id='search4'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'>
    <x xmlns='jabber:x:data' type='submit'>
      <field type='hidden' var='FORM_TYPE'>
        <value>jabber:iq:search</value>
      </field>
      <field var='x-gender'>
        <value>male</value>
      </field>
    </x>
  </query>
</iq>`},
	}

	return stanza.UnAckQueue{Uslice: q, TotalSent: len(q)}

}

func TestUnAckQueueAcknowledgeAndPushMonotonic(t *testing.T) {
	var uaq stanza.UnAckQueue
	_ = uaq.Push(&stanza.UnAckedStz{Stz: "<message id='1'/>"})
	_ = uaq.Push(&stanza.UnAckedStz{Stz: "<message id='2'/>"})
	_ = uaq.Push(&stanza.UnAckedStz{Stz: "<message id='3'/>"})

	if len(uaq.Uslice) != 3 || uaq.Uslice[2].Id != 3 {
		t.Fatalf("expected 3 items with last id 3, got len %d and id %d", len(uaq.Uslice), uaq.Uslice[2].Id)
	}

	// Ack all 3 items (queue becomes empty)
	acked := uaq.Acknowledge(3)
	if acked != 3 || len(uaq.Uslice) != 0 {
		t.Fatalf("expected 3 acked items and empty queue, got acked=%d len=%d", acked, len(uaq.Uslice))
	}

	// Push a 4th item after queue was emptied
	_ = uaq.Push(&stanza.UnAckedStz{Stz: "<message id='4'/>"})
	if len(uaq.Uslice) != 1 || uaq.Uslice[0].Id != 4 {
		t.Fatalf("expected 1 item with monotonic id 4, got len %d and id %d", len(uaq.Uslice), uaq.Uslice[0].Id)
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
