package work

type Worker interface {
    InSlot() <-chan transf.Transfer
    OutSlot() chan<- transf.Transfer
    BindSend(chan<- transf.Transfer) error
    BindState(chan<- transf.Transfer) error
    Run()
}

type Work struct {
    readCh          <-chan transf.Transfer
    writeCh         chan<- transf.Transfer
    sendCh          chan<- transf.Transfer
    stateCh         chan<- transf.Transfer
}

func (w *Work) InSlot() (<-chan transf.Transfer) {
    return w.writeCh
}

func (w *Work) OutSlot() (chan<- transf.Transfer) {
    return w.readCh
}

func (w *Work) BindSend(ch chan<- transf.Transfer) (error) {
    w.sendCh = ch
    return nil
}

func (w *Work) BindState(ch chan<- transf.Transfer) (error) {
    w.stateCh = ch
    return nil
}